package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	UpdateBulk(ctx context.Context, orders ...*domain.Order) error
	GetOrdersToTrade(context.Context, *domain.Order) ([]*domain.Order, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type TradeRepository interface {
	Save(ctx context.Context, trades ...*domain.Trade) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Trade, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Trade, error)
}

type MatchingEngine interface {
	ExecuteIncomingOrder(ctx context.Context, order *domain.Order) ([]*domain.Order, []*domain.Trade, error)
}
