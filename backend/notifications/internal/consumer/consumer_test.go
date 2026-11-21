package consumer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDecodeValidatesRecipientAndRetryRoute(t *testing.T) {
	user, id, order, security := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	key := "notifications.private." + user.String() + ".order.created"
	body, err := json.Marshal(map[string]any{"id": id, "type": key, "sequence": 1, "occurred_at": time.Now().UTC(), "payload": map[string]any{"schema_version": 1, "event_type": "order.created", "user_id": user, "order": map[string]any{"id": order, "security_id": security}}})
	if err != nil {
		t.Fatal(err)
	}
	d := amqp.Delivery{Body: body, RoutingKey: key, Type: key, MessageId: id.String()}
	n, err := decode(d, "queue")
	if err != nil || n.UserID != user || n.EventID != id {
		t.Fatalf("notification %+v: %v", n, err)
	}
	d.RoutingKey = "notifications.private." + uuid.NewString() + ".order.created"
	if _, err := decode(d, "queue"); err == nil {
		t.Fatal("accepted mismatched recipient")
	}
	d.RoutingKey = "queue"
	d.Headers = amqp.Table{retryHeader: int32(1)}
	if _, err := decode(d, "queue"); err != nil {
		t.Fatalf("valid retry rejected: %v", err)
	}
}
