package create_order

import (
	"context"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type Usecase struct {
	orderRepo      ports.OrderRepository
	orderGroupRepo ports.OrderGroupRepository
	securitySvc    ports.SecurityService
	userSvc        ports.UserService
}

func New(orderRepo ports.OrderRepository, orderGroupRepo ports.OrderGroupRepository, securitySvc ports.SecurityService, userSvc ports.UserService) *Usecase {
	return &Usecase{
		orderRepo:      orderRepo,
		orderGroupRepo: orderGroupRepo,
		securitySvc:    securitySvc,
		userSvc:        userSvc,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	// Validate trader exists
	if _, err := u.userSvc.GetByID(ctx, cmd.TraderID); err != nil {
		return Result{}, err
	}

	// Get security
	sec, err := u.securitySvc.GetByID(ctx, cmd.SecurityID)
	if err != nil {
		return Result{}, err
	}

	// Nominal value comes from quantity * security.nominal_value
	nominalValue := cmd.Quantity.Mul(sec.NominalValue)
	fx := sec.Fx

	// Create order group (aggregate root)
	orderGroup, err := domain.NewOrderGroup(cmd.TraderID, cmd.SecurityID, cmd.Direction, cmd.OrderType, cmd.Price, cmd.Yield, fx)
	if err != nil {
		return Result{}, err
	}

	if err := u.orderGroupRepo.Save(ctx, orderGroup); err != nil {
		return Result{}, err
	}

	order, err := domain.NewOrder(orderGroup.ID(), cmd.TraderID, cmd.SecurityID, cmd.Direction, cmd.OrderType, cmd.Price, cmd.Yield, cmd.Quantity, nominalValue, fx)
	if err != nil {
		return Result{}, err
	}

	if err := u.orderRepo.Save(ctx, order); err != nil {
		return Result{}, err
	}

	orderGroup.SetOrders([]*domain.Order{order})

	return Result{
		OrderGroupID: orderGroup.PublicID().String(),
		OrderID:      order.PublicID().String(),
		Status:       string(order.Status()),
	}, nil
}
