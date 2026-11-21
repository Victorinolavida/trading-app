package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
	"module/internal/adapters/out/postgres"
)

type retryStore struct {
	mu       sync.Mutex
	failures int
	saved    chan postgres.Notification
}

func (s *retryStore) Save(_ context.Context, n postgres.Notification) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failures > 0 {
		s.failures--
		return false, errors.New("temporary failure")
	}
	s.saved <- n
	return true, nil
}

func TestConsumerIntegrationRetryAndDeadLetter(t *testing.T) {
	if os.Getenv("NOTIFICATION_INTEGRATION") != "1" {
		t.Skip("set NOTIFICATION_INTEGRATION=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ch, err := client.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	exchange, queue := "notification.consumer."+suffix, "notification.consumer."+suffix
	defer func() {
		_, _ = ch.QueueDelete(queue, false, false, false)
		_, _ = ch.QueueDelete(queue+".dlq", false, false, false)
		_ = ch.ExchangeDelete(exchange, false, false)
	}()
	store := &retryStore{failures: 2, saved: make(chan postgres.Notification, 1)}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := ch.QueueDeclare(queue+".dlq", true, false, false, false, amqp.Table{"x-message-ttl": int32(604800000)}); err != nil {
		t.Fatal(err)
	}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": "", "x-dead-letter-routing-key": queue + ".dlq"}); err != nil {
		t.Fatal(err)
	}
	if err := ch.QueueBind(queue, "notifications.private.*.order.*", exchange, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := ch.ExchangeDeclare("notifications.live", "topic", true, false, false, false, nil); err != nil {
		t.Fatal(err)
	}
	live, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	c := New(client, store, exchange, queue)
	done := make(chan error, 1)
	go func() { done <- c.consume(ctx) }()
	user, id, order, security := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	key := "notifications.private." + user.String() + ".order.created"
	if err := ch.QueueBind(live.Name, user.String(), "notifications.live", false, nil); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"id": id, "type": key, "sequence": 1, "occurred_at": time.Now().UTC(), "payload": map[string]any{"schema_version": 1, "event_type": "order.created", "user_id": user, "order": map[string]any{"id": order, "security_id": security}}})
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{Body: body, Type: key, MessageId: id.String(), DeliveryMode: amqp.Persistent}); err != nil {
		t.Fatal(err)
	}
	var saved postgres.Notification
	select {
	case n := <-store.saved:
		saved = n
		if n.EventID != id {
			t.Fatalf("wrong event %s", n.EventID)
		}
	case <-ctx.Done():
		t.Fatal("retry did not succeed")
	}
	for {
		message, ok, err := ch.Get(live.Name, true)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			var signal struct {
				ID        uuid.UUID `json:"id"`
				EventType string    `json:"event_type"`
			}
			if err := json.Unmarshal(message.Body, &signal); err != nil || signal.ID != saved.ID || signal.EventType != "order.created" {
				t.Fatalf("invalid signal: %s (%v)", message.Body, err)
			}
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("live signal missing")
		case <-time.After(30 * time.Millisecond):
		}
	}
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{Body: []byte(`{}`), DeliveryMode: amqp.Persistent}); err != nil {
		t.Fatal(err)
	}
	for {
		_, ok, err := ch.Get(queue+".dlq", true)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("invalid event did not reach DLQ")
		case <-time.After(30 * time.Millisecond):
		}
	}
	store.mu.Lock()
	store.failures = 10
	store.mu.Unlock()
	if err := ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{Body: body, Type: key, MessageId: id.String(), DeliveryMode: amqp.Persistent}); err != nil {
		t.Fatal(err)
	}
	for {
		d, ok, err := ch.Get(queue+".dlq", true)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			if retryCount(d.Headers) != maxRetries {
				t.Fatalf("wrong retry count: %v", d.Headers)
			}
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("failed event did not reach DLQ")
		case <-time.After(30 * time.Millisecond):
		}
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("consumer did not stop")
	}
}
