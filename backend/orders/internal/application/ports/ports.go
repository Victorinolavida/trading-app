package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domain"
)

// ListOrdersFilter holds the optional filters of a trader order listing.
// Empty values mean "no filter".
type ListOrdersFilter struct {
	TraderID  uuid.UUID
	OrderIDs  []uuid.UUID
	Direction string
	Status    string
}

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	UpdateBulk(ctx context.Context, orders ...*domain.Order) error
	GetOrdersToTrade(context.Context, *domain.Order) ([]*domain.Order, error)
	GetActiveOrders(context.Context, uuid.UUID, uuid.UUID, string) ([]*domain.Order, error)
	ListOrders(context.Context, ListOrdersFilter) ([]*domain.Order, error)
	GetMarketDepth(context.Context, uuid.UUID) (*domain.MarketDepth, error)
	ExpireOrders(ctx context.Context) ([]*domain.Order, error)
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
	Lock(ctx context.Context, securityID string) error
}

type TradeRepository interface {
	Save(ctx context.Context, trades ...*domain.Trade) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Trade, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Trade, error)
}

type MatchingEngine interface {
	ExecuteIncomingOrder(ctx context.Context, order *domain.Order) ([]*domain.Order, []*domain.Trade, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, events ...domain.Event) error
}

type MemoryPort interface {
	Incr(context.Context, string) (int64, error)
}
