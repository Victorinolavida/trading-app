package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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

func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   d.Coefficient(),
		Exp:   d.Exponent(),
		Valid: true,
	}
}

func decimalNulltoNumeric(d decimal.NullDecimal) pgtype.Numeric {
	if !d.Valid {
		return pgtype.Numeric{Valid: false}
	}

	return decimalToNumeric(d.Decimal)
}

func (repo *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	return repo.queries.SaveOrder(ctx, sqlc.SaveOrderParams{
		ID:           order.ID(),
		PublicID:     order.PublicID(),
		TraderID:     order.TraderID(),
		SecurityID:   order.SecurityID(),
		Direction:    order.Direction().String(),
		OrderType:    order.OrderType().String(),
		Price:        decimalNulltoNumeric(order.Price()),
		Yield:        decimalNulltoNumeric(order.Yield()),
		Quantity:     int32(order.Quantity()),
		NominalValue: decimalToNumeric(order.NominalValue()),
		Fx:           decimalToNumeric(order.Fx()),
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

func numericToNullDecimal(num pgtype.Numeric) decimal.NullDecimal {
	if !num.Valid {
		return decimal.NullDecimal{}
	}
	return decimal.NewNullDecimal(decimal.NewFromBigInt(num.Int, num.Exp))
}

func toDomainOrder(o sqlc.Order) (*domain.Order, error) {
	price := numericToNullDecimal(o.Price)
	yield := numericToNullDecimal(o.Yield)
	nomValue := numericToNullDecimal(o.NominalValue)
	fx := numericToNullDecimal(o.Fx)
	allocationPtc := numericToNullDecimal(o.AllocationPtc)

	return domain.ReconstructOrder(
		o.ID,
		o.PublicID,
		o.TraderID,
		o.SecurityID,
		o.Direction,
		o.OrderType,
		o.Status,
		uint(o.Quantity),
		uint(o.AllocationQty),
		price,
		yield,
		nomValue.Decimal.InexactFloat64(),
		allocationPtc.Decimal.InexactFloat64(),
		fx.Decimal.InexactFloat64(),
		o.CreatedAt.Time,
		o.UpdatedAt.Time,
	)
}
