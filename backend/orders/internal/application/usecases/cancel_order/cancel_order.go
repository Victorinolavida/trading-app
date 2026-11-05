package cancel_order

import (
	"context"

	"github.com/victorinolavida/orders/internal/application/ports"
)

type Usecase struct {
	orderRepo ports.OrderRepository
}

func New(orderRepo ports.OrderRepository) *Usecase {
	return &Usecase{
		orderRepo: orderRepo,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	// Then cancel orders
	// for _, o := range orderGroup.Orders() {
	// 	if o.Status() == domain.OrderStatusActive {
	// 		if err := o.Cancel(); err != nil {
	// 			continue
	// 		}
	// 		if err := u.orderRepo.Update(ctx, o); err != nil {
	// 			return Result{}, err
	// 		}
	// 	}
	// }
	//
	// return Result{
	// 	OrderGroupID: orderGroup.PublicID().String(),
	// 	Status:       string(orderGroup.OrderGroupStatus()),
	// }, nil

	return Result{}, nil
}
