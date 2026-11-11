package cancel_order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Usecase struct {
	orderRepo ports.OrderRepository
	publisher ports.EventPublisher
}

func New(orderRepo ports.OrderRepository, publisher ports.EventPublisher) *Usecase {
	return &Usecase{
		orderRepo: orderRepo,
		publisher: publisher,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	if cmd.OrderID == uuid.Nil {
		return Result{}, domainerrors.ErrInvalidOrderId
	}

	order, err := u.orderRepo.GetByPublicID(ctx, cmd.OrderID)
	if err != nil {
		return Result{}, err
	}
	if cmd.TraderID != order.TraderID() {
		return Result{}, domainerrors.ErrOrderNotFound
	}

	if err := order.Cancel(); err != nil {
		if errors.Is(err, domainerrors.ErrOrderAlreadyCanceled) {
			return orderToResult(order), nil
		}
		return Result{}, err
	}

	if err := u.orderRepo.Update(ctx, order); err != nil {
		return Result{}, err
	}

	if u.publisher != nil {
		depth, err := u.orderRepo.GetMarketDepth(ctx, order.SecurityID())
		if err != nil {
			return Result{}, err
		}
		event := domain.NewMarketDepthEvent(order.SecurityID().String(), depth)
		if err := u.publisher.Publish(ctx, event); err != nil {
			return Result{}, err
		}
	}

	return orderToResult(order), nil
}

func orderToResult(order *domain.Order) Result {
	return Result{
		ID:                order.PublicID().String(),
		SecurityId:        order.SecurityID().String(),
		Direction:         order.Direction().String(),
		OrderType:         order.OrderType().String(),
		Price:             order.Price().String(),
		Yield:             order.Yield().String(),
		Quantity:          int64(order.Quantity()),
		NominalValue:      order.NominalValue().String(),
		AllocatedQuantity: int64(order.AllocationQty()),
		AllocationPtc:     order.AllocationPtc().String(),
		AllocationStatus:  order.AllocationStatus().String(),
		Fx:                order.Fx().String(),
		Status:            string(order.Status()),
		CreatedAt:         order.CreatedAt(),
		UpdatedAt:         order.UpdatedAt(),
	}
}
