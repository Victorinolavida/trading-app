package create_order

import (
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
	OrderGroupID string
	OrderID      string
	Status       string
}
