package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Order struct {
	id            uuid.UUID
	publicID      uuid.UUID
	traderID      uuid.UUID
	securityID    uuid.UUID
	direction     Direction
	orderType     OrderType
	price         decimal.NullDecimal
	yield         decimal.NullDecimal
	allocationPtc decimal.Decimal
	allocationQty uint
	quantity      uint
	nominalValue  decimal.Decimal
	fx            decimal.Decimal
	status        OrderStatus
	createdAt     time.Time
	updatedAt     time.Time
}

func (o *Order) ID() uuid.UUID                  { return o.id }
func (o *Order) PublicID() uuid.UUID            { return o.publicID }
func (o *Order) TraderID() uuid.UUID            { return o.traderID }
func (o *Order) SecurityID() uuid.UUID          { return o.securityID }
func (o *Order) Direction() Direction           { return o.direction }
func (o *Order) OrderType() OrderType           { return o.orderType }
func (o *Order) Price() decimal.NullDecimal     { return o.price }
func (o *Order) Yield() decimal.NullDecimal     { return o.yield }
func (o *Order) Quantity() uint                 { return o.quantity }
func (o *Order) NominalValue() decimal.Decimal  { return o.nominalValue }
func (o *Order) AllocationPtc() decimal.Decimal { return o.allocationPtc }
func (o *Order) AllocationQty() uint            { return o.allocationQty }
func (o *Order) Fx() decimal.Decimal            { return o.fx }
func (o *Order) Status() OrderStatus            { return o.status }
func (o *Order) CreatedAt() time.Time           { return o.createdAt }
func (o *Order) UpdatedAt() time.Time           { return o.updatedAt }
func (o *Order) touch()                         { o.updatedAt = time.Now() }

func NewOrder(
	traderID, securityID uuid.UUID,
	directionStr, orderTypeStr string, quantity uint,
	price, yield, nominalValue, fx decimal.Decimal,
) (*Order, error) {
	if traderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidTraderID
	}
	if securityID == uuid.Nil {
		return nil, domainerrors.ErrInvalidSecurityID
	}
	direction, err := NewDirection(directionStr)
	if err != nil {
		return nil, err
	}
	orderType, err := NewOrderType(orderTypeStr)
	if err != nil {
		return nil, err
	}
	if quantity == 0 {
		return nil, domainerrors.ErrInvalidQuantity
	}
	if nominalValue.LessThanOrEqual(decimal.Zero) {
		return nil, domainerrors.ErrInvalidNominalValue
	}
	if fx.LessThanOrEqual(decimal.Zero) {
		return nil, domainerrors.ErrInvalidFX
	}
	if price.LessThan(decimal.Zero) {
		return nil, domainerrors.ErrInvalidPrice
	}
	if yield.LessThan(decimal.Zero) {
		return nil, domainerrors.ErrInvalidYield
	}

	var nullablePrice, nullableYield decimal.NullDecimal

	if orderType == OrderTypeMarket {
		if price.IsPositive() {
			return nil, domainerrors.ErrPriceInMarketOrder
		}
		if yield.IsPositive() {
			return nil, domainerrors.ErrYieldInMarketOrder
		}
	} else {
		nullablePrice = decimal.NewNullDecimal(price)
		nullableYield = decimal.NewNullDecimal(yield)
	}

	now := time.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	publicID := uuid.New()

	return &Order{
		id:           id,
		publicID:     publicID,
		traderID:     traderID,
		securityID:   securityID,
		direction:    direction,
		orderType:    orderType,
		price:        nullablePrice,
		yield:        nullableYield,
		quantity:     quantity,
		nominalValue: nominalValue,
		fx:           fx,
		status:       OrderStatusActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func (o *Order) Cancel() error {
	if o.status == OrderStatusCancelled {
		return domainerrors.ErrOrderAlreadyAllocated
	}
	o.status = OrderStatusCancelled
	o.touch()
	return nil
}

func (o *Order) Kill() {
	o.status = OrderStatusKilled
	o.touch()
}

func (o *Order) Expire() {
	o.status = OrderStatusExpired
	o.touch()
}

func ReconstructOrder(
	id, publicID, traderID, securityID uuid.UUID,
	direction, orderType, status string, quantity, allocationQty uint,
	price, yiel decimal.NullDecimal, nominalValue, allocationPtc, fx float64,
	createdAt, updatedAt time.Time,
) (*Order, error) {
	dir, err := NewDirection(direction)
	if err != nil {
		return nil, err
	}
	ot, err := NewOrderType(orderType)
	if err != nil {
		return nil, err
	}
	st, err := NewOrderStatus(status)
	if err != nil {
		return nil, err
	}

	return &Order{
		id:            id,
		publicID:      publicID,
		traderID:      traderID,
		securityID:    securityID,
		allocationQty: allocationQty,
		allocationPtc: decimal.NewFromFloat(allocationPtc),
		direction:     dir,
		orderType:     ot,
		price:         price,
		yield:         yiel,
		quantity:      quantity,
		nominalValue:  decimal.NewFromFloat(nominalValue),
		fx:            decimal.NewFromFloat(fx),
		status:        st,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}
