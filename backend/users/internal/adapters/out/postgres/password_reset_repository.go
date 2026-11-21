package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/auth/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/auth/internal/application/ports"
)

type PasswordResetRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

var _ ports.PasswordResetRepository = (*PasswordResetRepository)(nil)

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool, queries: sqlc.New(pool)}
}

func (r *PasswordResetRepository) Store(ctx context.Context, userID uuid.UUID, digest []byte, expiresAt time.Time) error {
	return r.queries.StorePasswordResetToken(ctx, sqlc.StorePasswordResetTokenParams{
		UserID: userID, TokenHash: digest,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (r *PasswordResetRepository) Redeem(ctx context.Context, digest, passwordHash []byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := r.queries.WithTx(tx)
	userID, err := queries.ConsumePasswordResetToken(ctx, digest)
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrInvalidResetToken
	}
	if err != nil {
		return err
	}
	if err = queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{PasswordHash: string(passwordHash), ID: userID}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
