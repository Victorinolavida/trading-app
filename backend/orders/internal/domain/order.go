package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Order struct {
	id           uuid.UUID
	publicID     uuid.UUID
	orderGroupID uuid.UUID
	traderID     uuid.UUID
	securityID   uuid.UUID
	direction    Direction
	orderType    OrderType
	price        decimal.Decimal
	yield        decimal.Decimal
	quantity     decimal.Decimal
	nominalValue decimal.Decimal
	fx           decimal.Decimal
	status       OrderStatus
	createdAt    time.Time
	updatedAt    time.Time
}

func (o *Order) ID() uuid.UUID                { return o.id }
func (o *Order) PublicID() uuid.UUID          { return o.publicID }
func (o *Order) OrderGroupID() uuid.UUID      { return o.orderGroupID }
func (o *Order) TraderID() uuid.UUID          { return o.traderID }
func (o *Order) SecurityID() uuid.UUID        { return o.securityID }
func (o *Order) Direction() Direction         { return o.direction }
func (o *Order) OrderType() OrderType         { return o.orderType }
func (o *Order) Price() decimal.Decimal       { return o.price }
func (o *Order) Yield() decimal.Decimal       { return o.yield }
func (o *Order) Quantity() decimal.Decimal    { return o.quantity }
func (o *Order) NominalValue() decimal.Decimal { return o.nominalValue }
func (o *Order) Fx() decimal.Decimal          { return o.fx }
func (o *Order) Status() OrderStatus          { return o.status }
func (o *Order) CreatedAt() time.Time         { return o.createdAt }
func (o *Order) UpdatedAt() time.Time         { return o.updatedAt }
func (o *Order) touch()                      { o.updatedAt = time.Now() }

func NewOrder(
	orderGroupID, traderID, securityID uuid.UUID,
	directionStr, orderTypeStr string,
	price, yield, quantity, nominalValue, fx decimal.Decimal,
) (*Order, error) {
	if orderGroupID == uuid.Nil {
		return nil, domainerrors.ErrInvalidOrderGroupStatus
	}
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
	if quantity.LessThanOrEqual(decimal.Zero) {
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

	now := time.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	publicID := uuid.New()

	return &Order{
		id:           id,
		publicID:     publicID,
		orderGroupID: orderGroupID,
		traderID:     traderID,
		securityID:   securityID,
		direction:    direction,
		orderType:    orderType,
		price:        price,
		yield:        yield,
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
	id, publicID, orderGroupID, traderID, securityID uuid.UUID,
	direction, orderType, status string,
	price, yield, quantity, nominalValue, fx float64,
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
		id:           id,
		publicID:     publicID,
		orderGroupID: orderGroupID,
		traderID:     traderID,
		securityID:   securityID,
		direction:    dir,
		orderType:    ot,
		price:        decimal.NewFromFloat(price),
		yield:        decimal.NewFromFloat(yield),
		quantity:     decimal.NewFromFloat(quantity),
		nominalValue: decimal.NewFromFloat(nominalValue),
		fx:           decimal.NewFromFloat(fx),
		status:       st,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}
