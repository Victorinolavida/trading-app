package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type OrderRepository struct {
	queries *sqlc.Queries
}

var _ ports.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{queries: sqlc.New(pool)}
}

func (repo *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	return repo.queries.SaveOrder(ctx, sqlc.SaveOrderParams{
		ID:           order.ID(),
		PublicID:     order.PublicID(),
		OrderGroupID: order.OrderGroupID(),
		TraderID:     order.TraderID(),
		SecurityID:   order.SecurityID(),
		Direction:    order.Direction().String(),
		OrderType:    order.OrderType().String(),
		Price:        numericFromDecimal(order.Price()),
		Yield:        numericFromDecimal(order.Yield()),
		Quantity:     numericFromDecimal(order.Quantity()),
		NominalValue: numericFromDecimal(order.NominalValue()),
		Fx:           numericFromDecimal(order.Fx()),
		Status:       order.Status().String(),
		CreatedAt:    pgtype.Timestamptz{Time: order.CreatedAt(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: order.UpdatedAt(), Valid: true},
	})
}

func (repo *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	o, err := repo.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomainOrder(o)
}

func (repo *OrderRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.Order, error) {
	o, err := repo.queries.GetOrderByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomainOrder(o)
}

func (repo *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	return repo.queries.UpdateOrder(ctx, sqlc.UpdateOrderParams{
		Status:    order.Status().String(),
		UpdatedAt: pgtype.Timestamptz{Time: order.UpdatedAt(), Valid: true},
		ID:        order.ID(),
	})
}

func toDomainOrder(o sqlc.Order) (*domain.Order, error) {
	p, _ := decimalFromNumeric(o.Price)
	y, _ := decimalFromNumeric(o.Yield)
	q, _ := decimalFromNumeric(o.Quantity)
	nv, _ := decimalFromNumeric(o.NominalValue)
	fx, _ := decimalFromNumeric(o.Fx)

	return domain.ReconstructOrder(
		o.ID,
		o.PublicID,
		o.OrderGroupID,
		o.TraderID,
		o.SecurityID,
		o.Direction,
		o.OrderType,
		o.Status,
		p.InexactFloat64(),
		y.InexactFloat64(),
		q.InexactFloat64(),
		nv.InexactFloat64(),
		fx.InexactFloat64(),
		o.CreatedAt.Time,
		o.UpdatedAt.Time,
	)
}
