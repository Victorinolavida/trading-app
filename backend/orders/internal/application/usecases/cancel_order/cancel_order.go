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
	txManger  ports.TxManager
}

func New(orderRepo ports.OrderRepository, publisher ports.EventPublisher, manager ports.TxManager) *Usecase {
	return &Usecase{
		orderRepo: orderRepo,
		txManger:  manager,
		publisher: publisher,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	if cmd.OrderID == uuid.Nil {
		return Result{}, domainerrors.ErrInvalidOrderId
	}

	var orderResult *domain.Order
	err := u.txManger.WithTx(ctx, func(ctx context.Context) error {
		order, err := u.orderRepo.GetByPublicID(ctx, cmd.OrderID)
		if err != nil {
			return err
		}
		if err := u.txManger.Lock(ctx, order.SecurityID().String()); err != nil {
			return err
		}
		if cmd.TraderID != order.TraderID() {
			return domainerrors.ErrOrderNotFound
		}

		orderResult = order
		if err := order.Cancel(); err != nil {
			if errors.Is(err, domainerrors.ErrOrderAlreadyCanceled) {
				return nil
			}
			return err
		}

		if err := u.orderRepo.Update(ctx, order); err != nil {
			return err
		}

		if u.publisher != nil {
			depth, err := u.orderRepo.GetMarketDepth(ctx, order.SecurityID())
			if err != nil {
				return err
			}
			event := domain.NewMarketDepthEvent(order.SecurityID().String(), depth)
			if err := u.publisher.Publish(ctx, event); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}

	return orderToResult(orderResult), nil
}

func orderToResult(order *domain.Order) Result {
	return Result{
		ID:                order.PublicID().String(),
		SecurityId:        order.SecurityID().String(),
		Direction:         order.Direction().String(),
		OrderType:         order.OrderType().String(),
		Price:             domain.FormatDecimal(order.Price()),
		Yield:             domain.FormatDecimal(order.Yield()),
		Quantity:          int64(order.Quantity()),
		NominalValue:      order.NominalValue().String(),
		AllocatedQuantity: int64(order.AllocationQty()),
		AllocationPtc:     order.AllocationPtc().String(),
		AllocationStatus:  order.AllocationStatus().String(),
		Fx:                order.Fx().String(),
		Status:            string(order.Status()),
		CreatedAt:         order.CreatedAt(),
		UpdatedAt:         order.UpdatedAt(),
		ExpiresAt:         order.ExpiresAt(),
	}
}
