package create_order

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Command struct {
	TraderID   uuid.UUID
	SecurityID uuid.UUID
	Direction  string
	OrderType  string
	Price      decimal.Decimal
	Yield      decimal.Decimal
	Quantity   uint
}

type Result struct {
	OrderGroupID string
	OrderID      string
	Status       string
}
