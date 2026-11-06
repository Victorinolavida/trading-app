package get_order_group

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	ID           string
	OrderGroupID string
	TraderID     string
	SecurityID   string
	Direction    string
	OrderType    string
	Price        decimal.Decimal
	Yield        decimal.Decimal
	Quantity     decimal.Decimal
	NominalValue decimal.Decimal
	Fx           decimal.Decimal
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Result struct {
	ID               string
	PublicID         string
	TraderID         string
	SecurityID       string
	Direction        string
	OrderType        string
	AllocationPC     decimal.Decimal
	AllocationStatus string
	OrderGroupStatus string
	WeightedPrice    decimal.Decimal
	WeightedYield    decimal.Decimal
	WeightedFX       decimal.Decimal
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Orders           []Order
}
