package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type OrderRepository struct {
	queries *sqlc.Queries
}

func NewOrderRepository(pool *pgxpool.Pool) ports.OrderRepository {
	return &OrderRepository{queries: sqlc.New(pool)}
}

func (repo *OrderRepository) queriesFor(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return repo.queries.WithTx(tx)
	}
	return repo.queries
}

func ptrToInt8(p *domain.Decimal) pgtype.Int8 {
	if p == nil {
		return pgtype.Int8{} // Valid: false → NULL
	}
	return pgtype.Int8{Int64: p.Value(), Valid: true}
}

func int8ToDecimalPtr(n pgtype.Int8) *domain.Decimal {
	if !n.Valid {
		return nil
	}
	d := domain.Decimal(n.Int64)
	return &d
}

func (repo *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	return repo.queriesFor(ctx).SaveOrder(ctx, sqlc.SaveOrderParams{
		ID:               order.ID(),
		PublicID:         order.PublicID(),
		TraderID:         order.TraderID(),
		SecurityID:       order.SecurityID(),
		Direction:        order.Direction().String(),
		OrderType:        order.OrderType().String(),
		Price:            ptrToInt8(order.Price()),
		Yield:            ptrToInt8(order.Yield()),
		Quantity:         int64(order.Quantity()),
		AllocationQty:    int64(order.AllocationQty()),
		AllocationStatus: order.AllocationStatus().String(),
		NominalValue:     int64(order.NominalValue()),
		Fx:               int64(order.Fx()),
		Status:           order.Status().String(),
		CreatedAt:        pgtype.Timestamptz{Time: order.CreatedAt(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: order.UpdatedAt(), Valid: true},
	})
}

func (repo *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	o, err := repo.queriesFor(ctx).GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomainOrder(o)
}

func (repo *OrderRepository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.Order, error) {
	o, err := repo.queriesFor(ctx).GetOrderByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrOrderNotFound
		}
		return nil, err
	}
	return toDomainOrder(o)
}

func (repo *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	_, err := repo.queriesFor(ctx).UpdateOrder(ctx, sqlc.UpdateOrderParams{
		Status:           order.Status().String(),
		AllocationQty:    int64(order.AllocationQty()),
		AllocationStatus: order.AllocationStatus().String(),
		UpdatedAt:        pgtype.Timestamptz{Time: order.UpdatedAt(), Valid: true},
		ID:               order.ID(),
	})
	return err
}

func (repo *OrderRepository) UpdateBulk(ctx context.Context, orders ...*domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(orders))
	statuses := make([]string, 0, len(orders))
	allocationQtys := make([]int64, 0, len(orders))
	allocationStatuses := make([]string, 0, len(orders))
	updatedAts := make([]pgtype.Timestamptz, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID())
		statuses = append(statuses, order.Status().String())
		allocationQtys = append(allocationQtys, int64(order.AllocationQty()))
		allocationStatuses = append(allocationStatuses, order.AllocationStatus().String())
		updatedAts = append(updatedAts, pgtype.Timestamptz{Time: order.UpdatedAt(), Valid: true})
	}
	_, err := repo.queriesFor(ctx).UpdateOrders(ctx, sqlc.UpdateOrdersParams{
		Ids:                ids,
		Statuses:           statuses,
		AllocationQtys:     allocationQtys,
		AllocationStatuses: allocationStatuses,
		UpdatedAts:         updatedAts,
	})
	return err
}

func (repo *OrderRepository) GetOrdersToTrade(ctx context.Context, order *domain.Order) ([]*domain.Order, error) {
	var rows []sqlc.Order
	var err error
	if order.Direction() == domain.DirectionBid {
		rows, err = repo.queriesFor(ctx).GetAsksToTrade(ctx, order.SecurityID())
	} else {
		rows, err = repo.queriesFor(ctx).GetBidsToTrade(ctx, order.SecurityID())
	}
	if err != nil {
		return nil, err
	}
	orders := make([]*domain.Order, 0, len(rows))
	for _, row := range rows {
		o, err := toDomainOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (repo *OrderRepository) GetActiveOrders(ctx context.Context, traderID uuid.UUID, SecurityID uuid.UUID, direction string) ([]*domain.Order, error) {
	rows, err := repo.queries.GetActiveOrders(ctx, sqlc.GetActiveOrdersParams{
		TraderID:   traderID,
		SecurityID: SecurityID,
		Direction: pgtype.Text{
			String: direction,
			Valid:  len(direction) > 0,
		},
	})
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(rows))
	for _, row := range rows {
		o, err := toDomainOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (repo *OrderRepository) ListOrders(ctx context.Context, filter ports.ListOrdersFilter) ([]*domain.Order, error) {
	rows, err := repo.queries.ListOrders(ctx, sqlc.ListOrdersParams{
		TraderID: filter.TraderID,
		OrderIds: filter.OrderIDs,
		Direction: pgtype.Text{
			String: filter.Direction,
			Valid:  filter.Direction != "",
		},
		Status: pgtype.Text{
			String: filter.Status,
			Valid:  filter.Status != "",
		},
	})
	if err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(rows))
	for _, row := range rows {
		o, err := toDomainOrder(row)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func toDomainOrder(o sqlc.Order) (*domain.Order, error) {
	return domain.ReconstructOrder(
		o.ID,
		o.PublicID,
		o.TraderID,
		o.SecurityID,
		o.Direction,
		o.OrderType,
		o.Status,
		o.AllocationStatus,
		domain.Quantity(o.Quantity),
		domain.Quantity(o.AllocationQty),
		int8ToDecimalPtr(o.Price),
		int8ToDecimalPtr(o.Yield),
		domain.Decimal(o.NominalValue),
		domain.Decimal(o.Fx),
		o.CreatedAt.Time,
		o.UpdatedAt.Time,
	)
}
