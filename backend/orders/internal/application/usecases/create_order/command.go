package create_order

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	TraderID   uuid.UUID
	SecurityID uuid.UUID
	Direction  string
	OrderType  string
	Price      int64
	Yield      int64
	Quantity   int64
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
}
