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
}

type OrderGroupRepository interface {
	Save(ctx context.Context, orderGroup *domain.OrderGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderGroup, error)
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.OrderGroup, error)
	Update(ctx context.Context, orderGroup *domain.OrderGroup) error
	GetWithOrders(ctx context.Context, id uuid.UUID) (*domain.OrderGroup, error)
}

type MatchingEngine interface {
	ExecuteIncomingOrder(ctx context.Context, order *domain.OrderGroup) ([]*domain.Order, error)
}
