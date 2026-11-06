package domain

import (
	"errors"
	"slices"

	"github.com/victorinolavida/orders/internal/domainerrors"
)

type (
	Direction        string
	OrderType        string
	OrderStatus      string
	OrderGroupStatus string
	AllocationStatus string
)

const (
	DirectionBid Direction = "bid"
	DirectionAsk Direction = "ask"
)

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

const (
	OrderStatusActive         OrderStatus = "active"
	OrderStatusFullyAllocated OrderStatus = "fully_allocated"
	OrderStatusCancelled      OrderStatus = "cancelled"
	OrderStatusExpired        OrderStatus = "expired"
	OrderStatusKilled         OrderStatus = "killed"
)

const (
	AllocationStatusNone      AllocationStatus = "none"
	AllocationStatusPartial   AllocationStatus = "partial"
	AllocationStatusAllocated AllocationStatus = "allocated"
)

const (
	OrderGroupStatusActive         OrderGroupStatus = "active"
	OrderGroupStatusFullyAllocated OrderGroupStatus = "fully_allocated"
	OrderGroupStatusCancelled      OrderGroupStatus = "cancelled"
	OrderGroupStatusExpired        OrderGroupStatus = "expired"
	OrderGroupStatusKilled         OrderGroupStatus = "killed"
)

const (
	PriceScale            = 1_000 // three decimals
	DecimalScale          = 1_000_000
	MaxPrice              = 1_000 * PriceScale
	MaxQty       Quantity = 1_000_000_000
	// MaxPrice is already scaled by PriceScale, so the largest scaled
	// nominal value (price * qty) fits in int64: 1e6 * 1e9 = 1e15.
	MaxDecimalValue = int64(MaxPrice) * int64(MaxQty)
)

type (
	Quantity int64
	Decimal  int64
)

func (d Decimal) Value() int64 {
	return int64(d)
}

var (
	ErrInvalidNumericNumber    = errors.New("invalid numeric value")
	ErrInvalidAllocationStatus = errors.New("invalid allocation status")
)

func (d Direction) String() string { return string(d) }
func (d Direction) validate() error {
	if !slices.Contains([]Direction{DirectionBid, DirectionAsk}, d) {
		return domainerrors.ErrInvalidDirection
	}
	return nil
}

func NewDirection(val string) (Direction, error) {
	d := Direction(val)
	if err := d.validate(); err != nil {
		return Direction(""), err
	}
	return d, nil
}

func (o OrderType) String() string { return string(o) }
func (o OrderType) validate() error {
	if !slices.Contains([]OrderType{OrderTypeLimit, OrderTypeMarket}, o) {
		return domainerrors.ErrInvalidOrderType
	}
	return nil
}

func NewOrderType(val string) (OrderType, error) {
	o := OrderType(val)
	if err := o.validate(); err != nil {
		return OrderType(""), err
	}
	return o, nil
}

func (o OrderStatus) String() string { return string(o) }
func (o OrderStatus) validate() error {
	if !slices.Contains([]OrderStatus{OrderStatusActive, OrderStatusFullyAllocated, OrderStatusCancelled, OrderStatusExpired, OrderStatusKilled}, o) {
		return domainerrors.ErrInvalidOrderStatus
	}
	return nil
}

func NewOrderStatus(val string) (OrderStatus, error) {
	o := OrderStatus(val)
	if err := o.validate(); err != nil {
		return OrderStatus(""), err
	}
	return o, nil
}

func (o OrderGroupStatus) String() string { return string(o) }
func (o OrderGroupStatus) validate() error {
	if !slices.Contains([]OrderGroupStatus{OrderGroupStatusActive, OrderGroupStatusFullyAllocated, OrderGroupStatusCancelled, OrderGroupStatusExpired, OrderGroupStatusKilled}, o) {
		return domainerrors.ErrInvalidOrderGroupStatus
	}
	return nil
}

func NewOrderGroupStatus(val string) (OrderGroupStatus, error) {
	o := OrderGroupStatus(val)
	if err := o.validate(); err != nil {
		return OrderGroupStatus(""), err
	}
	return o, nil
}

func (a AllocationStatus) String() string { return string(a) }
func (a AllocationStatus) validate() error {
	if !slices.Contains([]AllocationStatus{AllocationStatusNone, AllocationStatusPartial, AllocationStatusAllocated}, a) {
		return ErrInvalidAllocationStatus
	}
	return nil
}

func NewAllocationStatus(val string) (AllocationStatus, error) {
	a := AllocationStatus(val)
	if err := a.validate(); err != nil {
		return AllocationStatus(""), err
	}
	return a, nil
}

func NewQuantity(qty int) (Quantity, error) {
	if qty <= 0 {
		return Quantity(0), domainerrors.ErrInvalidQuantity
	}
	q := Quantity(qty)
	if q > MaxQty {
		return Quantity(0), domainerrors.ErrQtyTooLarge
	}
	return q, nil
}

func NewDecimal(val int64) (Decimal, error) {
	if val < 0 || val > MaxDecimalValue {
		return Decimal(0), ErrInvalidNumericNumber
	}
	return Decimal(val), nil
}
