package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/victorinolavida/auth/internal/adapters/secondary_adapter/postgres/sqlc"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	queries *sqlc.Queries
}

var _ ports.UserRepository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: sqlc.New(pool)}
}

func (repo *UserRepository) Save(ctx context.Context, user domain.User) error {
	return repo.queries.SaveUser(ctx, sqlc.SaveUserParams{
		ID:           user.ID(),
		Email:        user.Email().String(),
		FirstName:    user.FirstName().String(),
		LastName:     user.LastName().String(),
		PasswordHash: string(user.Hash()),
		CreatedAt:    pgtype.Timestamptz{Time: user.CreatedAt(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: user.UpdatedAt(), Valid: true},
	})
}

func (repo *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	user, err := repo.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domainerrors.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(user)
}

func (repo *UserRepository) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	user, err := repo.queries.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domainerrors.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomain(user)
}

func (repo *UserRepository) ExistByEmail(ctx context.Context, email domain.Email) (bool, error) {
	_, err := repo.queries.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func toDomain(u sqlc.User) (domain.User, error) {
	var createdAt, updatedAt time.Time
	if u.CreatedAt.Valid {
		createdAt = u.CreatedAt.Time
	}
	if u.UpdatedAt.Valid {
		updatedAt = u.UpdatedAt.Time
	}
	return domain.Reconstruct(
		u.ID,
		u.Email,
		u.FirstName,
		u.LastName,
		[]byte(u.PasswordHash),
		createdAt,
		updatedAt,
	)
}

func (repo *UserRepository) Update(ctx context.Context, user domain.User) error {
	return repo.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		FirstName: user.FirstName().String(),
		UpdatedAt: pgtype.Timestamptz{Time: user.UpdatedAt(), Valid: true},
		LastName:  user.LastName().String(),
		ID:        user.ID(),
	})
}
