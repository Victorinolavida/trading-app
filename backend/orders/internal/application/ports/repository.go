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

type MatchingEngine interface {
	ExecuteIncomingOrder(ctx context.Context, order *domain.Order) ([]*domain.Order, error)
}
