package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/orders/internal/application/ports"
)

type txContextKey struct{}

var (
	ErrLockNotAcquired = errors.New("security lock not acquired")
	ErrNoTx            = errors.New("security lock requires an active transaction")
)

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx, ok
}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) ports.TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(withTx(ctx, tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *TxManager) Lock(ctx context.Context, securityID string) error {
	tx, ok := txFromContext(ctx)
	if !ok {
		return ErrNoTx
	}

	if err := sqlc.New(tx).LockSecurity(ctx, securityID); err != nil {
		return fmt.Errorf("lock security %s: %w", securityID, err)
	}
	return nil
}
