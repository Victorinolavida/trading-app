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

type TradeRepository struct {
	queries *sqlc.Queries
}

var _ ports.TradeRepository = (*TradeRepository)(nil)

func NewTradeRepository(pool *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{queries: sqlc.New(pool)}
}

func (repo *TradeRepository) queriesFor(ctx context.Context) *sqlc.Queries {
	if tx, ok := txFromContext(ctx); ok {
		return repo.queries.WithTx(tx)
	}
	return repo.queries
}

func (repo *TradeRepository) Save(ctx context.Context, trades ...*domain.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(trades))
	publicIDs := make([]uuid.UUID, 0, len(trades))
	makerOrderIDs := make([]uuid.UUID, 0, len(trades))
	takerOrderIDs := make([]uuid.UUID, 0, len(trades))
	prices := make([]int64, 0, len(trades))
	quantities := make([]int64, 0, len(trades))
	takerSides := make([]string, 0, len(trades))
	executedAts := make([]pgtype.Timestamptz, 0, len(trades))
	for _, trade := range trades {
		ids = append(ids, trade.ID)
		publicIDs = append(publicIDs, trade.PublicID)
		makerOrderIDs = append(makerOrderIDs, trade.MakerOrderID)
		takerOrderIDs = append(takerOrderIDs, trade.TakerOrderID)
		prices = append(prices, int64(trade.Price))
		quantities = append(quantities, int64(trade.Quantity))
		takerSides = append(takerSides, trade.TakerSide.String())
		executedAts = append(executedAts, pgtype.Timestamptz{Time: trade.ExecutedAt, Valid: true})
	}
	_, err := repo.queriesFor(ctx).SaveTrades(ctx, sqlc.SaveTradesParams{
		Ids:           ids,
		PublicIds:     publicIDs,
		MakerOrderIds: makerOrderIDs,
		TakerOrderIds: takerOrderIDs,
		Prices:        prices,
		Quantities:    quantities,
		TakerSides:    takerSides,
		ExecutedAts:   executedAts,
	})
	return err
}

func (repo *TradeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Trade, error) {
	t, err := repo.queriesFor(ctx).GetTradeByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrTradeNotFound
		}
		return nil, err
	}
	return toDomainTrade(t), nil
}

func (repo *TradeRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.Trade, error) {
	rows, err := repo.queriesFor(ctx).GetTradesByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	trades := make([]*domain.Trade, 0, len(rows))
	for _, row := range rows {
		trades = append(trades, toDomainTrade(row))
	}
	return trades, nil
}

func toDomainTrade(t sqlc.Trade) *domain.Trade {
	return &domain.Trade{
		ID:           t.ID,
		PublicID:     t.PublicID,
		MakerOrderID: t.MakerOrderID,
		TakerOrderID: t.TakerOrderID,
		Price:        domain.Decimal(t.Price),
		Quantity:     domain.Quantity(t.Quantity),
		TakerSide:    domain.Direction(t.TakerSide),
		ExecutedAt:   t.ExecutedAt.Time,
	}
}
