package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/broker"
)

var ErrNotConfirmed = errors.New("rabbitmq: message not confirmed by broker")

// Publisher sends messages to a topic exchange with a routing key.
// Each Publisher holds its own AMQP channel.
type Publisher[T broker.Eventable] struct {
	channel  *amqp.Channel
	exchange string
}

func NewPublisher[T broker.Eventable](client *Client, exchange string) (*Publisher[T], error) {
	ch, err := client.Channel()
	if err != nil {
		return nil, err
	}

	if err := ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		return nil, err
	}

	if err := ch.Confirm(false); err != nil {
		return nil, err
	}

	return &Publisher[T]{channel: ch, exchange: exchange}, nil
}

// Publish sends a message to the exchange with the given routing key.
// Waits for broker confirmation before returning.
func (p *Publisher[T]) Publish(ctx context.Context, payload T, sequence int64) error {
	event := broker.NewEvent(payload, sequence)
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("rabbitmq: marshal event %s: %w", event.ID, err)
	}

	return p.publishRaw(ctx, event.Type, event.Type, event.ID.String(), event.OccurredAt, body)
}

func (p *Publisher[T]) publishRaw(
	ctx context.Context,
	routingKey, eventType, messageID string,
	ts time.Time,
	body []byte,
) error {
	conf, err := p.channel.PublishWithDeferredConfirmWithContext(ctx,
		p.exchange,
		routingKey,
		false, // mandatory
		false, // immediate (not supported by RabbitMQ)
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    messageID,
			Type:         eventType, // logical type, e.g. "order.created"
			Timestamp:    ts,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("rabbitmq: publish %s: %w", routingKey, err)
	}

	acked, err := conf.WaitContext(ctx)
	if err != nil {
		return fmt.Errorf("rabbitmq: wait confirm %s: %w", routingKey, err)
	}
	if !acked {
		return ErrNotConfirmed
	}

	slog.Debug("rabbitmq: published", slog.String("exchange", p.exchange), slog.String("routingKey", routingKey), slog.String("id", messageID))
	return nil
}

func (p *Publisher[T]) Close() error {
	return p.channel.Close()
}
