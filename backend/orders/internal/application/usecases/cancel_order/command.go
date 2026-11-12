package cancel_order

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	OrderID  uuid.UUID
	TraderID uuid.UUID
}

type Result struct {
	ID                string
	SecurityId        string
	Direction         string
	OrderType         string
	Price             string
	Yield             string
	Quantity          int64
	NominalValue      string
	AllocatedQuantity int64
	AllocationPtc     string
	AllocationStatus  string
	Fx                string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         time.Time
}
