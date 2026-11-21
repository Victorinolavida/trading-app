package create_order

import (
	"context"
	"log/slog"
	"sort"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type Usecase struct {
	orderRepo      ports.OrderRepository
	tradeRepo      ports.TradeRepository
	securitySvc    ports.SecurityService
	userSvc        ports.UserService
	matchingEngine ports.MatchingEngine
	txManager      ports.TxManager
	publisher      ports.EventPublisher
	calculator     Calculator
	outbox         ports.NotificationOutbox
}

func New(orderRepo ports.OrderRepository, tradeRepo ports.TradeRepository, securitySvc ports.SecurityService, userSvc ports.UserService, me ports.MatchingEngine, tx ports.TxManager, publisher ports.EventPublisher, calculator Calculator, outbox ports.NotificationOutbox) *Usecase {
	return &Usecase{
		orderRepo:      orderRepo,
		tradeRepo:      tradeRepo,
		securitySvc:    securitySvc,
		userSvc:        userSvc,
		matchingEngine: me,
		txManager:      tx,
		publisher:      publisher,
		calculator:     calculator,
		outbox:         outbox,
	}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	slog.Debug("incoming order received",
		slog.String("trader_id", cmd.TraderID.String()),
		slog.String("security_id", cmd.SecurityID.String()),
		slog.String("direction", cmd.Direction),
		slog.String("order_type", cmd.OrderType),
		slog.Int64("quantity", cmd.Quantity),
		slog.Int64("price", cmd.Price),
		slog.Int64("yield", cmd.Yield),
	)

	// Validate trader exists
	if _, err := u.userSvc.GetByID(ctx, cmd.TraderID); err != nil {
		return Result{}, err
	}

	// Get security
	sec, err := u.securitySvc.GetByID(ctx, cmd.SecurityID)
	if err != nil {
		return Result{}, err
	}

	if err := u.completeLimitOrder(ctx, &cmd); err != nil {
		return Result{}, err
	}

	// Nominal value comes from quantity * security.nominal_value
	qty := domain.Quantity(cmd.Quantity)
	if qty != 0 && sec.NominalValue > domain.Decimal(domain.MaxDecimalValue)/domain.Decimal(qty) {
		return Result{}, domain.ErrInvalidNumericNumber
	}
	nominalValue := sec.NominalValue * domain.Decimal(qty)

	order, err := domain.NewOrder(
		cmd.TraderID, cmd.SecurityID,
		cmd.Direction, cmd.OrderType,
		int(cmd.Quantity),
		cmd.Price, cmd.Yield,
		nominalValue, sec.Fx, cmd.ExpiresAt,
	)
	if err != nil {
		return Result{}, err
	}

	if err := u.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := u.txManager.Lock(ctx, order.SecurityID().String()); err != nil {
			return err
		}

		if err := u.orderRepo.Save(ctx, order); err != nil {
			return err
		}
		// Keep the creation snapshot before matching changes the order.
		created := *order

		matches, trades, internalCrossing, err := u.matchingEngine.ExecuteIncomingOrder(ctx, order)
		if err != nil {
			return err
		}

		var matchedQty domain.Quantity
		for _, trade := range trades {
			matchedQty += trade.Quantity
		}
		slog.Debug("order matching result",
			slog.String("order_id", order.ID().String()),
			slog.Bool("matched", len(trades) > 0),
			slog.Int("trades", len(trades)),
			slog.Int64("matched_quantity", int64(matchedQty)),
		)

		if len(trades) > 0 || !order.IsActive() {
			orders := make([]*domain.Order, 0, len(matches)+1)
			orders = append(orders, order)
			orders = append(orders, matches...)
			if err := u.orderRepo.UpdateBulk(ctx, orders...); err != nil {
				return err
			}
			if len(trades) > 0 {
				if err := u.tradeRepo.Save(ctx, trades...); err != nil {
					return err
				}
			}
		}
		type notification struct {
			kind   string
			order  *domain.Order
			reason string
		}
		notifications := []notification{{"order.created", &created, ""}}
		for _, match := range matches {
			notifications = append(notifications, notification{"order.executed", order, ""}, notification{"order.executed", match, ""})
		}
		if internalCrossing {
			notifications = append(notifications, notification{"order.discarded", order, "internal_crossing"})
		}
		// Acquire recipient sequence locks in a consistent order across securities.
		sort.SliceStable(notifications, func(i, j int) bool {
			return notifications[i].order.TraderID().String() < notifications[j].order.TraderID().String()
		})
		for _, n := range notifications {
			if err := u.outbox.EnqueueOrder(ctx, n.kind, n.order, n.reason); err != nil {
				return err
			}
		}
		depth, err := u.orderRepo.GetMarketDepth(ctx, order.SecurityID())
		if err != nil {
			return err
		}
		if err := u.outbox.EnqueueMarketDepth(ctx, domain.NewMarketDepthEvent(order.SecurityID().String(), depth)); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return Result{}, err
	}

	return Result{
		ID:                order.PublicID().String(),
		SecurityId:        order.SecurityID().String(),
		Direction:         order.Direction().String(),
		OrderType:         order.OrderType().String(),
		Price:             domain.FormatDecimal(order.Price()),
		Yield:             domain.FormatDecimal(order.Yield()),
		Quantity:          order.Quantity().Value(),
		NominalValue:      order.NominalValue().String(),
		AllocatedQuantity: order.AllocationQty().Value(),
		AllocationPtc:     order.AllocationPtc().String(),
		AllocationStatus:  order.AllocationStatus().String(),
		Fx:                order.Fx().String(),
		Status:            order.Status().String(),
		CreatedAt:         order.CreatedAt(),
		UpdatedAt:         order.UpdatedAt(),
		ExpiresAt:         order.ExpiresAt(),
	}, nil
}
