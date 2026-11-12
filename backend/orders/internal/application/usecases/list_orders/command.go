package list_orders

import (
	"time"

	"github.com/google/uuid"
)

type Command struct {
	TraderID  uuid.UUID
	OrderIDs  []uuid.UUID
	Direction string
	Status    string
}

type Order struct {
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

type Result struct {
	Orders []Order
}
