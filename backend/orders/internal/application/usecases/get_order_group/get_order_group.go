package get_order_group

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
)

type Usecase struct {
	orderGroupRepo ports.OrderGroupRepository
}

func New(orderGroupRepo ports.OrderGroupRepository) *Usecase {
	return &Usecase{orderGroupRepo: orderGroupRepo}
}

func (u *Usecase) Execute(ctx context.Context, id uuid.UUID) (Result, error) {
	og, err := u.orderGroupRepo.GetWithOrders(ctx, id)
	if err != nil {
		return Result{}, err
	}

	res := Result{
		ID:               og.ID().String(),
		PublicID:         og.PublicID().String(),
		TraderID:         og.TraderID().String(),
		SecurityID:       og.SecurityID().String(),
		Direction:        string(og.Direction()),
		OrderType:        string(og.OrderType()),
		AllocationPC:     og.AllocationPC(),
		AllocationStatus: string(og.AllocationStatus()),
		OrderGroupStatus: string(og.OrderGroupStatus()),
		WeightedPrice:    og.WeightedPrice(),
		WeightedYield:    og.WeightedYield(),
		WeightedFX:       og.WeightedFX(),
		CreatedAt:        og.CreatedAt(),
		UpdatedAt:        og.UpdatedAt(),
		Orders:           make([]Order, 0, len(og.Orders())),
	}

	for _, o := range og.Orders() {
		res.Orders = append(res.Orders, Order{
			ID:           o.ID().String(),
			OrderGroupID: o.OrderGroupID().String(),
			TraderID:     o.TraderID().String(),
			SecurityID:   o.SecurityID().String(),
			Direction:    string(o.Direction()),
			OrderType:    string(o.OrderType()),
			Price:        o.Price(),
			Yield:        o.Yield(),
			Quantity:     o.Quantity(),
			NominalValue: o.NominalValue(),
			Fx:           o.Fx(),
			Status:       string(o.Status()),
			CreatedAt:    o.CreatedAt(),
			UpdatedAt:    o.UpdatedAt(),
		})
	}

	return res, nil
}
