package matchingengine

import (
	"context"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type MatchingEngineService struct {
	repo ports.OrderRepository
}

func New(repo ports.OrderRepository) ports.MatchingEngine {
	return &MatchingEngineService{repo}
}

func (me *MatchingEngineService) ExecuteIncomingOrder(ctx context.Context, incomingOrder *domain.Order) ([]*domain.Order, []*domain.Trade, error) {
	ordersToTrade, err := me.repo.GetOrdersToTrade(ctx, incomingOrder)
	if err != nil {
		return nil, nil, err
	}
	matches := make([]*domain.Order, 0, len(ordersToTrade))
	trades := make([]*domain.Trade, 0, len(ordersToTrade))

	for _, order := range ordersToTrade {
		execution, validMatch := order.MatchWith(incomingOrder)
		if !validMatch {
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
	return matches, trades, nil
}
