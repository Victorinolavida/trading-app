package domain

import (
	"errors"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
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
	OrderGroupStatusActive    OrderGroupStatus = "active"
	OrderGroupStatusAllocated OrderGroupStatus = "allocated"
	OrderGroupStatusCancelled OrderGroupStatus = "cancelled"
	OrderGroupStatusExpired   OrderGroupStatus = "expired"
)

const (
	AllocationStatusPending   AllocationStatus = "pending"
	AllocationStatusPartial   AllocationStatus = "partial"
	AllocationStatusAllocated AllocationStatus = "allocated"
)

var ErrInvalidNumericNumber = errors.New("invalid numeric value")

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
	if !slices.Contains([]OrderGroupStatus{OrderGroupStatusActive, OrderGroupStatusAllocated, OrderGroupStatusCancelled, OrderGroupStatusExpired}, o) {
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
	if !slices.Contains([]AllocationStatus{AllocationStatusPending, AllocationStatusPartial, AllocationStatusAllocated}, a) {
		return domainerrors.ErrInvalidOrderGroupStatus
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

func NumericToFloat(n pgtype.Numeric) (float64, error) {
	val, err := n.Float64Value()
	if err != nil {
		return float64(0), err
	}
	if !val.Valid {
		return float64(0), ErrInvalidNumericNumber
	}
	return val.Float64, nil
}
