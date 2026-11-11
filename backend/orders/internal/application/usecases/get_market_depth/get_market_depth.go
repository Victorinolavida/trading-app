package get_market_depth

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Usecase struct {
	orderRepo ports.OrderRepository
}

func New(orderRepo ports.OrderRepository) *Usecase {
	return &Usecase{
		orderRepo: orderRepo,
	}
}

type Result struct {
	MarketDepth *domain.MarketDepth
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	if cmd.SecurityID == uuid.Nil {
		return Result{}, domainerrors.ErrInvalidSecurityID
	}

	depth, err := u.orderRepo.GetMarketDepth(ctx, cmd.SecurityID)
	if err != nil {
		return Result{}, err
	}

	return Result{
		MarketDepth: depth,
	}, nil
}
