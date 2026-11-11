package rabbitmq

import (
	"context"
	"encoding/json"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/broker"
	"github.com/victorinolavida/pkg/config"
)

// HandlerFunc processes a single event. Return an error to Nack and requeue,
// return nil to Ack.
type HandlerFunc[T broker.Eventable] func(ctx context.Context, event broker.Event[T]) error

// Consumer reads messages from a queue bound to a topic exchange.
// Each Consumer holds its own AMQP channel.
type Consumer[T broker.Eventable] struct {
	channel *amqp.Channel
	queue   string
}

func NewConsumer[T broker.Eventable](
	client *Client,
	exchange, queue, routingKey string,
) (*Consumer[T], error) {
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

	if _, err := ch.QueueDeclare(
		queue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	); err != nil {
		return nil, err
	}

	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return nil, err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return nil, err
	}

	return &Consumer[T]{channel: ch, queue: queue}, nil
}

func NewConsumerFromConfig[T broker.Eventable](client *Client, config config.Consumer) (*Consumer[T], error) {
	return NewConsumer[T](client, config.Exchange, config.Queue, config.RoutingKey)
}

// Consume starts consuming messages and calls handler for each one.
// Blocks until ctx is cancelled.
func (c *Consumer[T]) Consume(ctx context.Context, handler HandlerFunc[T]) error {
	deliveries, err := c.channel.Consume(
		c.queue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	slog.Info("rabbitmq: consuming from queue", slog.String("queue", c.queue))
	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-deliveries:
			if !ok {
				slog.Warn("rabbitmq: delivery channel closed")
				return nil
			}
			var event broker.Event[T]
			if err := json.Unmarshal(d.Body, &event); err != nil {
				slog.Error("rabbitmq: unmarshal event", slog.String("queue", c.queue), slog.Any("error", err))
				d.Nack(false, false)
				continue
			}
			if err := handler(ctx, event); err != nil {
				slog.Error("rabbitmq: handler error, requeueing", slog.Any("error", err))
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}
}

func (c *Consumer[T]) Close() error {
	return c.channel.Close()
}
