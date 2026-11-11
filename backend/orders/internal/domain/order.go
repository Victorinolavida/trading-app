package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Order struct {
	id               uuid.UUID
	publicID         uuid.UUID
	traderID         uuid.UUID
	securityID       uuid.UUID
	direction        Direction
	orderType        OrderType
	price            *Decimal
	yield            *Decimal
	allocationQty    Quantity
	allocationStatus AllocationStatus
	quantity         Quantity
	nominalValue     Decimal
	fx               Decimal
	status           OrderStatus
	createdAt        time.Time
	updatedAt        time.Time
}

func (o *Order) ID() uuid.UUID                      { return o.id }
func (o *Order) PublicID() uuid.UUID                { return o.publicID }
func (o *Order) TraderID() uuid.UUID                { return o.traderID }
func (o *Order) SecurityID() uuid.UUID              { return o.securityID }
func (o *Order) Direction() Direction               { return o.direction }
func (o *Order) OrderType() OrderType               { return o.orderType }
func (o *Order) Price() *Decimal                    { return o.price }
func (o *Order) Yield() *Decimal                    { return o.yield }
func (o *Order) Quantity() Quantity                 { return o.quantity }
func (o *Order) NominalValue() Decimal              { return o.nominalValue }
func (o *Order) AllocationStatus() AllocationStatus { return o.allocationStatus }
func (o *Order) AllocationQty() Quantity            { return o.allocationQty }
func (o *Order) Fx() Decimal                        { return o.fx }
func (o *Order) Status() OrderStatus                { return o.status }
func (o *Order) CreatedAt() time.Time               { return o.createdAt }
func (o *Order) UpdatedAt() time.Time               { return o.updatedAt }
func (o *Order) touch()                             { o.updatedAt = time.Now() }
func (o *Order) AllocationPtc() Decimal {
	if o.quantity <= 0 || o.allocationQty <= 0 {
		return Decimal(0)
	}
	return Decimal(o.allocationQty.Value() * 100 * int64(DecimalScale) / o.quantity.Value())
}

func toDecimalPtr(v int64) (*Decimal, error) {
	if v == 0 {
		return nil, nil // absent: market order or unset yield
	}
	d, err := NewDecimal(v)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func NewOrder(
	traderID, securityID uuid.UUID,
	directionStr, orderTypeStr string, quantity int,
	price, yield int64, nominalValue, fx Decimal,
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
	qty, err := NewQuantity(quantity)
	if err != nil {
		return nil, err
	}

	if nominalValue <= 0 {
		return nil, domainerrors.ErrInvalidNominalValue
	}
	if fx <= 0 {
		return nil, domainerrors.ErrInvalidFX
	}
	if err := validateDecimalRange(nominalValue); err != nil {
		return nil, err
	}
	if err := validateDecimalRange(fx); err != nil {
		return nil, err
	}
	if price < 0 {
		return nil, domainerrors.ErrInvalidPrice
	}
	if yield < 0 {
		return nil, domainerrors.ErrInvalidYield
	}

	pricePtr, err := toDecimalPtr(price)
	if err != nil {
		return nil, err
	}
	yieldPtr, err := toDecimalPtr(yield)
	if err != nil {
		return nil, err
	}

	if orderType == OrderTypeMarket {
		if pricePtr != nil {
			return nil, domainerrors.ErrPriceInMarketOrder
		}
		if yieldPtr != nil {
			return nil, domainerrors.ErrYieldInMarketOrder
		}
	} else if pricePtr == nil {
		return nil, domainerrors.ErrInvalidPrice
	}

	now := time.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	publicID := uuid.New()

	return &Order{
		id:               id,
		publicID:         publicID,
		traderID:         traderID,
		securityID:       securityID,
		direction:        direction,
		orderType:        orderType,
		price:            pricePtr,
		yield:            yieldPtr,
		quantity:         qty,
		nominalValue:     nominalValue,
		fx:               fx,
		allocationQty:    0,
		allocationStatus: AllocationStatusNone,
		status:           OrderStatusActive,
		createdAt:        now,
		updatedAt:        now,
	}, nil
}

func validateDecimalRange(d Decimal) error {
	if d < 0 || d > Decimal(MaxDecimalValue) {
		return ErrInvalidNumericNumber
	}
	return nil
}

func (o *Order) Cancel() error {
	if o.status == OrderStatusCancelled {
		return domainerrors.ErrOrderAlreadyCanceled
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
	direction, orderType, status, allocationStatus string,
	quantity, allocationQty Quantity,
	price, yield *Decimal, nominalValue, fx Decimal,
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
	as, err := NewAllocationStatus(allocationStatus)
	if err != nil {
		return nil, err
	}
	if quantity <= 0 || quantity > MaxQty {
		return nil, ErrInvalidNumericNumber
	}
	if allocationQty < 0 || allocationQty > quantity {
		return nil, ErrInvalidNumericNumber
	}

	return &Order{
		id:               id,
		publicID:         publicID,
		traderID:         traderID,
		securityID:       securityID,
		allocationQty:    allocationQty,
		allocationStatus: as,
		direction:        dir,
		orderType:        ot,
		price:            price,
		yield:            yield,
		quantity:         quantity,
		nominalValue:     nominalValue,
		fx:               fx,
		status:           st,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
	}, nil
}

func (o *Order) IsMarket() bool { return o.orderType == OrderTypeMarket }
func (o *Order) IsLimit() bool  { return o.orderType == OrderTypeLimit }

type Execution struct {
	Qty   Quantity
	Price Decimal
	Taker Direction
}

// Remaining return the remaining/active quantity of a order
func (o *Order) Remaining() Quantity {
	return o.quantity - o.allocationQty
}

func (o *Order) execute(e Execution) {
	o.allocationQty += e.Qty
	o.updateAllocationStatus()
	o.touch()
}

func (o *Order) updateAllocationStatus() {
	switch {
	case o.allocationQty == 0:
		o.allocationStatus = AllocationStatusNone
	case o.allocationQty >= o.quantity:
		o.allocationStatus = AllocationStatusAllocated
		o.status = OrderStatusFullyAllocated
	default:
		o.allocationStatus = AllocationStatusPartial
	}
}

func (o *Order) IsCancelled() bool {
	return o.status == OrderStatusCancelled
}

func (o *Order) IsExpired() bool {
	return o.status == OrderStatusExpired
}

func (o *Order) IsActive() bool {
	return o.status == OrderStatusActive
}

func (o *Order) canMatch(maker *Order) bool {
	if o.Remaining() == 0 || maker.Remaining() == 0 {
		return false
	}
	if o.status == OrderStatusCancelled || o.status == OrderStatusExpired {
		return false
	}
	if maker.status == OrderStatusCancelled || maker.status == OrderStatusExpired {
		return false
	}
	if o.direction == maker.direction {
		return false
	}
	return true
}

func (o *Order) MatchWith(maker *Order) (Execution, bool) {
	if !o.canMatch(maker) {
		return Execution{}, false
	}
	qty := min(o.Remaining(), maker.Remaining())
	// Execution price starts from the maker's price; a market maker
	// has none of its own, so it takes the taker's price. If both
	// are market there is no price to discover.
	var price Decimal
	if maker.price != nil {
		price = *maker.price
	}
	if o.price != nil {
		price = *o.price
	}
	if price == 0 {
		return Execution{}, false
	}
	execution := Execution{
		Qty:   qty,
		Price: price,
		Taker: o.direction,
	}
	o.execute(execution)
	maker.execute(execution)
	return execution, true
}
