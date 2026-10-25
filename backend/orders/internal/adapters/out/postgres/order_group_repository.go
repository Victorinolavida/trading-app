package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type OrderGroupRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

var _ ports.OrderGroupRepository = (*OrderGroupRepository)(nil)

func NewOrderGroupRepository(pool *pgxpool.Pool) *OrderGroupRepository {
	return &OrderGroupRepository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (repo *OrderGroupRepository) Save(ctx context.Context, orderGroup *domain.OrderGroup) error {
	return repo.queries.SaveOrderGroup(ctx, sqlc.SaveOrderGroupParams{
		ID:               orderGroup.ID(),
		PublicID:         orderGroup.PublicID(),
		TraderID:         orderGroup.TraderID(),
		SecurityID:       orderGroup.SecurityID(),
		Direction:        orderGroup.Direction().String(),
		OrderType:        orderGroup.OrderType().String(),
		AllocationPc:     numericFromDecimal(orderGroup.AllocationPC()),
		AllocationStatus: orderGroup.AllocationStatus().String(),
		OrderGroupStatus: orderGroup.OrderGroupStatus().String(),
		WeightedPrice:    numericFromDecimal(orderGroup.WeightedPrice()),
		WeightedYield:    numericFromDecimal(orderGroup.WeightedYield()),
		WeightedFx:       numericFromDecimal(orderGroup.WeightedFX()),
		CreatedAt:        pgtype.Timestamptz{Time: orderGroup.CreatedAt(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: orderGroup.UpdatedAt(), Valid: true},
	})
}

func (repo *OrderGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderGroup, error) {
	og, err := repo.queries.GetOrderGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderGroupNotFound
		}
		return nil, err
	}
	return toDomainOrderGroup(og)
}

func (repo *OrderGroupRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.OrderGroup, error) {
	og, err := repo.queries.GetOrderGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderGroupNotFound
		}
		return nil, err
	}
	return toDomainOrderGroup(og)
}

func (repo *OrderGroupRepository) Update(ctx context.Context, orderGroup *domain.OrderGroup) error {
	return repo.queries.UpdateOrderGroup(ctx, sqlc.UpdateOrderGroupParams{
		AllocationPc:     numericFromDecimal(orderGroup.AllocationPC()),
		AllocationStatus: orderGroup.AllocationStatus().String(),
		OrderGroupStatus: orderGroup.OrderGroupStatus().String(),
		WeightedPrice:    numericFromDecimal(orderGroup.WeightedPrice()),
		WeightedYield:    numericFromDecimal(orderGroup.WeightedYield()),
		WeightedFx:       numericFromDecimal(orderGroup.WeightedFX()),
		UpdatedAt:        pgtype.Timestamptz{Time: orderGroup.UpdatedAt(), Valid: true},
		ID:               orderGroup.ID(),
	})
}

func (repo *OrderGroupRepository) GetWithOrders(ctx context.Context, id uuid.UUID) (*domain.OrderGroup, error) {
	og, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	orders, err := repo.queries.GetOrdersByOrderGroupID(ctx, id)
	if err != nil {
		return nil, err
	}
	domainOrders := make([]*domain.Order, 0, len(orders))
	for _, o := range orders {
		do, err := toDomainOrder(o)
		if err != nil {
			return nil, err
		}
		domainOrders = append(domainOrders, do)
	}
	og.SetOrders(domainOrders)
	return og, nil
}

func toDomainOrderGroup(og sqlc.OrderGroup) (*domain.OrderGroup, error) {
	apc, _ := decimalFromNumeric(og.AllocationPc)
	wp, _ := decimalFromNumeric(og.WeightedPrice)
	wy, _ := decimalFromNumeric(og.WeightedYield)
	wfx, _ := decimalFromNumeric(og.WeightedFx)

	var createdAt, updatedAt pgtype.Timestamptz
	if og.CreatedAt.Valid {
		createdAt = og.CreatedAt
	}
	if og.UpdatedAt.Valid {
		updatedAt = og.UpdatedAt
	}

	return domain.ReconstructOrderGroup(
		og.ID,
		og.PublicID,
		og.TraderID,
		og.SecurityID,
		og.Direction,
		og.OrderType,
		og.AllocationStatus,
		og.OrderGroupStatus,
		apc.InexactFloat64(),
		wp.InexactFloat64(),
		wy.InexactFloat64(),
		wfx.InexactFloat64(),
		createdAt.Time,
		updatedAt.Time,
	)
}

func numericFromDecimal(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	s := d.String()
	_ = n.Scan(s)
	return n
}

func decimalFromNumeric(n pgtype.Numeric) (decimal.Decimal, error) {
	f, err := domain.NumericToFloat(n)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(f), nil
}
