package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Trade struct {
	ID           uuid.UUID
	PublicID     uuid.UUID
	MakerOrderID uuid.UUID
	TakerOrderID uuid.UUID
	Price        Decimal // execution price, resting order's price
	Quantity     Quantity
	TakerSide    Direction // which takerSide was the aggressor, useful for market data
	ExecutedAt   time.Time
}

func NewTrade(makerOrderID, takerOrderID uuid.UUID, price Decimal, qty Quantity, takerSide Direction) (*Trade, error) {
	if makerOrderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidMakerID
	}
	if takerOrderID == uuid.Nil {
		return nil, domainerrors.ErrInvalidMakerID
	}
	if qty <= 0 {
		return nil, domainerrors.ErrInvalidQuantity
	}
	if price <= 0 {
		return nil, domainerrors.ErrInvalidPrice
	}
	if err := takerSide.validate(); err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	return &Trade{
		ID:           id,
		PublicID:     uuid.New(),
		MakerOrderID: makerOrderID,
		TakerOrderID: takerOrderID,
		Price:        price,
		Quantity:     qty,
		TakerSide:    takerSide,
		ExecutedAt:   time.Now(),
	}, nil
}
