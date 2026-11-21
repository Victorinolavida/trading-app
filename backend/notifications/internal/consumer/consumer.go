package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
	"module/internal/adapters/out/postgres"
)

const maxRetries int32 = 3
const retryHeader = "notification-retries"

type Store interface {
	Save(context.Context, postgres.Notification) (bool, error)
}

type Consumer struct {
	client          *rabbitmq.Client
	store           Store
	exchange, queue string
}

func New(client *rabbitmq.Client, store Store, exchange, queue string) *Consumer {
	return &Consumer{client: client, store: store, exchange: exchange, queue: queue}
}

// Run recreates channels after connection loss; unacknowledged deliveries return to the queue.
func (c *Consumer) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := c.consume(ctx); err != nil && ctx.Err() == nil {
			slog.Error("notification consumer stopped", slog.Any("error", err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(c.exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare("notifications.live", "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(c.queue+".dlq", true, false, false, false, amqp.Table{"x-message-ttl": int32(604800000)}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(c.queue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": c.queue + ".dlq"}); err != nil {
		return err
	}
	if err := ch.QueueBind(c.queue, "notifications.private.*.order.*", c.exchange, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(1, 0, false); err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return err
	}
	returns := ch.NotifyReturn(make(chan amqp.Return, 1))
	deliveries, err := ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			n, err := decode(d, c.queue)
			if err != nil {
				slog.Warn("invalid notification event", slog.Any("error", err))
				if err := d.Reject(false); err != nil {
					return err
				}
				continue
			}
			if inserted, saveErr := c.store.Save(ctx, n); saveErr == nil {
				if inserted {
					// This signal is best effort; clients reconcile from the database on reconnect.
					confirmation, signalErr := ch.PublishWithDeferredConfirmWithContext(ctx, "notifications.live", n.UserID.String(), false, false, amqp.Publishing{ContentType: "application/json", Body: []byte(`{}`)})
					if signalErr == nil {
						_, signalErr = confirmation.WaitContext(ctx)
					}
					if signalErr != nil {
						slog.Warn("notification live signal failed", slog.Any("error", signalErr))
					}
				}
				if err = d.Ack(false); err != nil {
					return err
				}
				continue
			} else {
				err = saveErr
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			slog.Error("saving notification failed", slog.Any("error", err))
			retries := retryCount(d.Headers)
			if retries >= maxRetries {
				if err := d.Reject(false); err != nil {
					return err
				}
				continue
			}
			headers := amqp.Table{}
			for k, v := range d.Headers {
				headers[k] = v
			}
			headers[retryHeader] = retries + 1
			conf, err := ch.PublishWithDeferredConfirmWithContext(ctx, "", c.queue, true, false, amqp.Publishing{Headers: headers, ContentType: d.ContentType, DeliveryMode: amqp.Persistent, MessageId: d.MessageId, Type: d.Type, Timestamp: d.Timestamp, Body: d.Body})
			if err != nil {
				return err
			}
			ack, err := conf.WaitContext(ctx)
			if err != nil {
				return err
			}
			if !ack {
				return errors.New("retry was not confirmed")
			}
			select {
			case ret := <-returns:
				return fmt.Errorf("retry unroutable: %s", ret.ReplyText)
			default:
			}
			if err := d.Ack(false); err != nil {
				return err
			}
		}
	}
}

func retryCount(headers amqp.Table) int32 {
	if n, ok := headers[retryHeader].(int32); ok && n >= 0 {
		return n
	}
	return 0
}

func decode(d amqp.Delivery, queue string) (postgres.Notification, error) {
	var event struct {
		ID         uuid.UUID       `json:"id"`
		Type       string          `json:"type"`
		Sequence   int64           `json:"sequence"`
		OccurredAt time.Time       `json:"occurred_at"`
		Payload    json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(d.Body, &event); err != nil {
		return postgres.Notification{}, err
	}
	var payload struct {
		SchemaVersion int       `json:"schema_version"`
		EventType     string    `json:"event_type"`
		UserID        uuid.UUID `json:"user_id"`
		Order         struct {
			ID         uuid.UUID `json:"id"`
			SecurityID uuid.UUID `json:"security_id"`
		} `json:"order"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return postgres.Notification{}, err
	}
	switch payload.EventType {
	case "order.created", "order.canceled", "order.expired", "order.executed", "order.discarded":
	default:
		return postgres.Notification{}, fmt.Errorf("unknown event type %q", payload.EventType)
	}
	key := "notifications.private." + payload.UserID.String() + "." + payload.EventType
	validRoute := d.RoutingKey == key || (d.RoutingKey == queue && retryCount(d.Headers) > 0)
	if event.ID == uuid.Nil || payload.UserID == uuid.Nil || payload.Order.ID == uuid.Nil || payload.Order.SecurityID == uuid.Nil || event.Sequence <= 0 || event.OccurredAt.IsZero() || payload.SchemaVersion != 1 || event.Type != key || !validRoute || (d.Type != "" && d.Type != key) || (d.MessageId != "" && d.MessageId != event.ID.String()) || !strings.HasPrefix(key, "notifications.private.") {
		return postgres.Notification{}, errors.New("invalid notification event envelope or recipient")
	}
	return postgres.Notification{ID: uuid.New(), UserID: payload.UserID, EventID: event.ID, EventType: payload.EventType, Payload: event.Payload, CreatedAt: event.OccurredAt}, nil
}
