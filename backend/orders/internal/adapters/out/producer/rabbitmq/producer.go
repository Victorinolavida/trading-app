package rabbitmq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	broker "github.com/victorinolavida/pkg/broker/rabbitmq"
)

type Producer struct {
	redis     ports.MemoryPort
	publisher *broker.Publisher[*domain.MarketDepthEvent]
}

func NewProducer(redis ports.MemoryPort, client *broker.Client, exchange string) (ports.EventPublisher, error) {
	publisher, err := broker.NewPublisher[*domain.MarketDepthEvent](client, exchange)
	if err != nil {
		return nil, err
	}

	return &Producer{
		redis:     redis,
		publisher: publisher,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, events ...domain.Event) error {
	if len(events) == 0 {
		slog.Warn("no events to send")
		return nil
	}

	for _, event := range events {
		md, ok := event.(*domain.MarketDepthEvent)
		if !ok {
			return fmt.Errorf("rabbitmq: unsupported event type %T", event)
		}

		sequence, err := p.redis.Incr(ctx, md.EventType())
		if err != nil {
			return fmt.Errorf("rabbitmq: increment sequence for %s: %w", md.EventType(), err)
		}

		if err := p.publisher.Publish(ctx, md, sequence); err != nil {
			return err
		}
	}

	return nil
}

func (p *Producer) Close() error {
	return p.publisher.Close()
}
