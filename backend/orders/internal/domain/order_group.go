package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type OrderGroup struct {
	id                uuid.UUID
	publicID          uuid.UUID
	traderID          uuid.UUID
	securityID        uuid.UUID
	direction         Direction
	orderType         OrderType
	allocationPC      decimal.Decimal
	allocationStatus  AllocationStatus
	orderGroupStatus  OrderGroupStatus
	weightedPrice     decimal.Decimal
	weightedYield     decimal.Decimal
	weightedFX        decimal.Decimal
	createdAt         time.Time
	updatedAt         time.Time
	orders            []*Order
}

func (og *OrderGroup) ID() uuid.UUID                        { return og.id }
func (og *OrderGroup) PublicID() uuid.UUID                  { return og.publicID }
func (og *OrderGroup) TraderID() uuid.UUID                  { return og.traderID }
func (og *OrderGroup) SecurityID() uuid.UUID                { return og.securityID }
func (og *OrderGroup) Direction() Direction                 { return og.direction }
func (og *OrderGroup) OrderType() OrderType                 { return og.orderType }
func (og *OrderGroup) AllocationPC() decimal.Decimal        { return og.allocationPC }
func (og *OrderGroup) AllocationStatus() AllocationStatus   { return og.allocationStatus }
func (og *OrderGroup) OrderGroupStatus() OrderGroupStatus   { return og.orderGroupStatus }
func (og *OrderGroup) WeightedPrice() decimal.Decimal       { return og.weightedPrice }
func (og *OrderGroup) WeightedYield() decimal.Decimal       { return og.weightedYield }
func (og *OrderGroup) WeightedFX() decimal.Decimal          { return og.weightedFX }
func (og *OrderGroup) CreatedAt() time.Time                 { return og.createdAt }
func (og *OrderGroup) UpdatedAt() time.Time                 { return og.updatedAt }
func (og *OrderGroup) Orders() []*Order                     { return og.orders }
func (og *OrderGroup) SetOrders(orders []*Order)            { og.orders = orders }
func (og *OrderGroup) touch()                               { og.updatedAt = time.Now() }

func NewOrderGroup(
	traderID, securityID uuid.UUID,
	directionStr, orderTypeStr string,
	weightedPrice, weightedYield, weightedFX decimal.Decimal,
) (*OrderGroup, error) {
	direction, err := NewDirection(directionStr)
	if err != nil {
		return nil, err
	}
	orderType, err := NewOrderType(orderTypeStr)
	if err != nil {
		return nil, err
	}
	if traderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidTraderID
	}
	if securityID == uuid.Nil {
		return nil, domainerrors.ErrInvalidSecurityID
	}

	now := time.Now()
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	publicID := uuid.New()

	return &OrderGroup{
		id:               id,
		publicID:         publicID,
		traderID:         traderID,
		securityID:       securityID,
		direction:        direction,
		orderType:        orderType,
		allocationPC:     decimal.Zero,
		allocationStatus: AllocationStatusPending,
		orderGroupStatus: OrderGroupStatusActive,
		weightedPrice:    weightedPrice,
		weightedYield:    weightedYield,
		weightedFX:       weightedFX,
		createdAt:        now,
		updatedAt:        now,
		orders:           []*Order{},
	}, nil
}

func (og *OrderGroup) Cancel() error {
	if og.orderGroupStatus == OrderGroupStatusCancelled {
		return domainerrors.ErrOrderGroupAlreadyCanceled
	}
	og.orderGroupStatus = OrderGroupStatusCancelled
	og.touch()
	return nil
}

func (og *OrderGroup) Activate() {
	og.orderGroupStatus = OrderGroupStatusActive
	og.touch()
}

func ReconstructOrderGroup(
	id, publicID, traderID, securityID uuid.UUID,
	direction, orderType, allocationStatus, orderGroupStatus string,
	allocationPC, weightedPrice, weightedYield, weightedFX float64,
	createdAt, updatedAt time.Time,
) (*OrderGroup, error) {
	dir, err := NewDirection(direction)
	if err != nil {
		return nil, err
	}
	ot, err := NewOrderType(orderType)
	if err != nil {
		return nil, err
	}
	as, err := NewAllocationStatus(allocationStatus)
	if err != nil {
		return nil, err
	}
	ogs, err := NewOrderGroupStatus(orderGroupStatus)
	if err != nil {
		return nil, err
	}

	return &OrderGroup{
		id:               id,
		publicID:         publicID,
		traderID:         traderID,
		securityID:       securityID,
		direction:        dir,
		orderType:        ot,
		allocationPC:     decimal.NewFromFloat(allocationPC),
		allocationStatus: as,
		orderGroupStatus: ogs,
		weightedPrice:    decimal.NewFromFloat(weightedPrice),
		weightedYield:    decimal.NewFromFloat(weightedYield),
		weightedFX:       decimal.NewFromFloat(weightedFX),
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		orders:           []*Order{},
	}, nil
}
