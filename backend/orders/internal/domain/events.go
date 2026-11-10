package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	OrderBookSnapshot EventType = "order-book-snapshot"
)

type Event struct {
	ID         uuid.UUID
	Type       EventType
	SecurityID uuid.UUID
	Sequence   int64 // per-security monotonic, for gap detection
	OccurredAt time.Time
	Payload    any
}
