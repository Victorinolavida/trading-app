package matchingengine

import (
	"context"
	"log/slog"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type MatchingEngineService struct {
	repo          ports.OrderRepository
	cossingPolicy domain.InternalCrossingPolicy
}

func New(repo ports.OrderRepository, crossingPolicy domain.InternalCrossingPolicy) ports.MatchingEngine {
	return &MatchingEngineService{repo, crossingPolicy}
}

func (me *MatchingEngineService) canMatch(takerOrder, makerOrder *domain.Order) (string, bool) {
	executionResult := canExecute(takerOrder, makerOrder)
	if executionResult != MatchOK {
		return executionResult.String(), false
	}

	action := me.cossingPolicy.Decide(takerOrder, makerOrder)
	if action != domain.CrossingAllow {
		return RejectInternalCrossing.String(), false
	}

	return "", true
}

func (me *MatchingEngineService) ExecuteIncomingOrder(ctx context.Context, incomingOrder *domain.Order) ([]*domain.Order, []*domain.Trade, error) {
	ordersToTrade, err := me.repo.GetOrdersToTrade(ctx, incomingOrder)
	if err != nil {
		return nil, nil, err
	}
	matches := make([]*domain.Order, 0, len(ordersToTrade))
	trades := make([]*domain.Trade, 0, len(ordersToTrade))

	for _, order := range ordersToTrade {
		rejectReason, ok := me.canMatch(incomingOrder, order)
		if !ok {
			slog.Debug(
				"order was rejected",
				slog.String("reason", rejectReason),
				slog.String("order_id", order.ID().String()),
			)
			killUnfilledMarketOrder(incomingOrder)
			return matches, trades, nil
		}

		execution, validMatch := incomingOrder.MatchWith(order)
		if !validMatch {
			killUnfilledMarketOrder(incomingOrder)
			return matches, trades, nil
		}

		trade, err := domain.NewTrade(
			order.ID(),
			incomingOrder.ID(),
			execution.Price,
			execution.Qty,
			incomingOrder.Direction(),
		)
		if err != nil {
			return nil, nil, err
		}

		matches = append(matches, order)
		trades = append(trades, trade)

	}
	killUnfilledMarketOrder(incomingOrder)
	return matches, trades, nil
}

// A market order only cares about quantity: whatever is left unfilled
// after matching has no price to rest at, so it is killed.
func killUnfilledMarketOrder(order *domain.Order) {
	if order.IsMarket() && order.Remaining() > 0 {
		order.Kill()
	}
}
