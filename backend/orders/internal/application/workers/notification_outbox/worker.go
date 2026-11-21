package notificationoutbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
)

const queueName = "notifications.orders"

type Worker struct {
	pool     *pgxpool.Pool
	client   *rabbitmq.Client
	exchange string
	queue    string
}

func New(pool *pgxpool.Pool, client *rabbitmq.Client, exchange string) *Worker {
	return &Worker{pool: pool, client: client, exchange: exchange, queue: queueName}
}

// Run retries failures, including channel failures after a broker reconnect.
func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := w.drain(ctx); err != nil && ctx.Err() == nil {
			slog.Error("notification outbox publisher failed", slog.Any("error", err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (w *Worker) drain(ctx context.Context) error {
	ch, err := w.client.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := declareTopology(ch, w.exchange, w.queue); err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return err
	}
	returns := ch.NotifyReturn(make(chan amqp.Return, 1))
	for ctx.Err() == nil {
		found, err := w.publishOne(ctx, ch, returns)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
	}
	return ctx.Err()
}

func declareTopology(ch *amqp.Channel, exchange, queue string) error {
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	// Dead-lettering through the default exchange sends rejected messages to the DLQ.
	if _, err := ch.QueueDeclare(queue+".dlq", true, false, false, false, amqp.Table{"x-message-ttl": int32(604800000)}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": queue + ".dlq"}); err != nil {
		return err
	}
	return ch.QueueBind(queue, "notifications.private.*.order.*", exchange, false, nil)
}

func (w *Worker) publishOne(ctx context.Context, ch *amqp.Channel, returns <-chan amqp.Return) (bool, error) {
	tx, err := w.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	var id uuid.UUID
	var sequence int64
	var key string
	var occurred time.Time
	var payload []byte
	err = tx.QueryRow(ctx, `SELECT id, sequence, routing_key, occurred_at, payload FROM notification_outbox WHERE published_at IS NULL ORDER BY occurred_at, id FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &sequence, &key, &occurred, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	body, err := marshalEvent(id, key, sequence, occurred, payload)
	if err != nil {
		return false, err
	}
	confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, w.exchange, key, true, false, amqp.Publishing{ContentType: "application/json", DeliveryMode: amqp.Persistent, MessageId: id.String(), Type: key, Timestamp: occurred, Body: body})
	if err != nil {
		return false, err
	}
	acked, err := confirmation.WaitContext(ctx)
	if err != nil {
		return false, err
	}
	if !acked {
		return false, errors.New("notification publish was not confirmed")
	}
	select {
	case returned := <-returns:
		return false, fmt.Errorf("notification unroutable: %s", returned.ReplyText)
	default:
	}
	if _, err := tx.Exec(ctx, `UPDATE notification_outbox SET published_at = now() WHERE id = $1`, id); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

func marshalEvent(id uuid.UUID, key string, sequence int64, occurred time.Time, payload json.RawMessage) ([]byte, error) {
	return json.Marshal(struct {
		ID         uuid.UUID       `json:"id"`
		Type       string          `json:"type"`
		Sequence   int64           `json:"sequence"`
		OccurredAt time.Time       `json:"occurred_at"`
		Payload    json.RawMessage `json:"payload"`
	}{id, key, sequence, occurred, payload})
}
