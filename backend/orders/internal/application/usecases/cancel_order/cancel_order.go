package cancel_order

import (
	"context"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Usecase struct {
	orderRepo      ports.OrderRepository
	orderGroupRepo ports.OrderGroupRepository
}

func New(orderRepo ports.OrderRepository, orderGroupRepo ports.OrderGroupRepository) *Usecase {
	return &Usecase{
		orderRepo:      orderRepo,
		orderGroupRepo: orderGroupRepo,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	orderGroup, err := u.orderGroupRepo.GetWithOrders(ctx, cmd.OrderGroupID)
	if err != nil {
		return Result{}, err
	}
	if orderGroup.TraderID() != cmd.TraderID {
		return Result{}, domainerrors.ErrOrderGroupNotFound
	}

	// Cancel order group first
	if err := orderGroup.Cancel(); err != nil {
		return Result{}, err
	}
	if err := u.orderGroupRepo.Update(ctx, orderGroup); err != nil {
		return Result{}, err
	}

	// Then cancel orders
	for _, o := range orderGroup.Orders() {
		if o.Status() == domain.OrderStatusActive {
			if err := o.Cancel(); err != nil {
				continue
			}
			if err := u.orderRepo.Update(ctx, o); err != nil {
				return Result{}, err
			}
		}
	}

	return Result{
		OrderGroupID: orderGroup.PublicID().String(),
		Status:       string(orderGroup.OrderGroupStatus()),
	}, nil
}
