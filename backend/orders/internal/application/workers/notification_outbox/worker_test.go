package notificationoutbox

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMarshalEventPreservesEnvelopeAndPayload(t *testing.T) {
	id := uuid.New()
	key := "notifications.private.user.order.created"
	timestamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	body, err := marshalEvent(id, key, 42, timestamp, json.RawMessage(`{"event_type":"order.created","quantity":5}`))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		ID         uuid.UUID `json:"id"`
		Type       string    `json:"type"`
		Sequence   int64     `json:"sequence"`
		OccurredAt time.Time `json:"occurred_at"`
		Payload    struct {
			EventType string `json:"event_type"`
			Quantity  int    `json:"quantity"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != id || got.Type != key || got.Sequence != 42 || !got.OccurredAt.Equal(timestamp) || got.Payload.EventType != "order.created" || got.Payload.Quantity != 5 {
		t.Fatalf("unexpected envelope: %+v", got)
	}
}
