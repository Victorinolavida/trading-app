package create_order

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type Usecase struct {
	orderRepo   ports.OrderRepository
	securitySvc ports.SecurityService
	userSvc     ports.UserService
}

func New(orderRepo ports.OrderRepository, securitySvc ports.SecurityService, userSvc ports.UserService) *Usecase {
	return &Usecase{
		orderRepo:   orderRepo,
		securitySvc: securitySvc,
		userSvc:     userSvc,
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
	nominalValue := sec.NominalValue.Mul(decimal.NewFromUint64(uint64(cmd.Quantity)))
	fx := sec.Fx

	order, err := domain.NewOrder(
		cmd.TraderID, cmd.SecurityID,
		cmd.Direction, cmd.OrderType, cmd.Quantity,
		cmd.Price, cmd.Yield,
		nominalValue, fx,
	)
	if err != nil {
		return Result{}, err
	}

	if err := u.orderRepo.Save(ctx, order); err != nil {
		return Result{}, err
	}

	return Result{
		OrderID: order.PublicID().String(),
		Status:  string(order.Status()),
	}, nil
}
