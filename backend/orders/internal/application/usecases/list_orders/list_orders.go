package list_orders

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

func orderToResult(order *domain.Order) Order {
	return Order{
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

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	if cmd.TraderID == uuid.Nil {
		return Result{}, domainerrors.ErrInvalidTraderID
	}
	if cmd.Direction != "" {
		if _, err := domain.NewDirection(cmd.Direction); err != nil {
			return Result{}, err
		}
	}
	if cmd.Status != "" {
		if _, err := domain.NewOrderStatus(cmd.Status); err != nil {
			return Result{}, err
		}
	}

	orders, err := u.orderRepo.ListOrders(ctx, ports.ListOrdersFilter{
		TraderID:  cmd.TraderID,
		OrderIDs:  cmd.OrderIDs,
		Direction: cmd.Direction,
		Status:    cmd.Status,
	})
	if err != nil {
		return Result{}, err
	}

	ordersResult := make([]Order, len(orders))
	for i, o := range orders {
		ordersResult[i] = orderToResult(o)
	}

	return Result{
		Orders: ordersResult,
	}, nil
}
