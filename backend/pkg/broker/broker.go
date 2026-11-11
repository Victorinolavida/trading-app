package broker

import (
	"time"

	"github.com/google/uuid"
)

type Eventable interface {
	EventType() string // e.g. "order.created"
}

type Event[T Eventable] struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Sequence   int64     `json:"sequence"` // per-security monotonic, for gap detection
	OccurredAt time.Time `json:"occurred_at"`
	Payload    T         `json:"payload"`
}

func NewEvent[T Eventable](payload T, sequence int64) Event[T] {
	return Event[T]{
		ID:         uuid.New(),
		Type:       payload.EventType(),
		Sequence:   sequence,
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}
}
