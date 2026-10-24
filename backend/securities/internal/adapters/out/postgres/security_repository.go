package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/securities/internal/adapters/out/postgres/sqlc"
	"github.com/victorinolavida/securities/internal/application/ports"
	"github.com/victorinolavida/securities/internal/domain"
	"github.com/victorinolavida/securities/internal/domainerrors"
)

type SecurityRepo struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

func NewSecurityRepo(pool *pgxpool.Pool) *SecurityRepo {
	return &SecurityRepo{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

var _ ports.SecurityRepository = (*SecurityRepo)(nil)

func (repo *SecurityRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Security, error) {
	sec, err := repo.queries.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrSecurityNotFound
		}
		return nil, err
	}

	return toDomain(sec)
}

func (repo *SecurityRepo) GetByISIN(ctx context.Context, isin string) (*domain.Security, error) {
	sec, err := repo.queries.GetByISIN(ctx, isin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
			return nil, domainerrors.ErrSecurityNotFound
		}
		return nil, err
	}

	return toDomain(sec)
}

func (repo *SecurityRepo) GetListByParams(ctx context.Context, filters ports.Params) ([]*domain.Security, error) {
	qb := sq.Select(
		"id", "public_id", "isin", "name", "type", "serie", "issue_date", "maturity_date",
		"nominal_value", "currency", "fx", "enable", "created_at", "updated_at",
	).
		From("securities").
		OrderBy("issue_date DESC").
		PlaceholderFormat(sq.Dollar)

	if len(filters.IDs) > 0 {
		qb = qb.Where(sq.Expr("public_id = ANY(?)", filters.IDs))
	}
	if len(filters.ISINs) > 0 {
		qb = qb.Where(sq.Expr("isin = ANY(?)", filters.ISINs))
	}
	if len(filters.Serie) > 0 {
		qb = qb.Where(sq.Expr("serie = ANY(?)", filters.Serie))
	}
	if len(filters.SecurityTypes) > 0 {
		types := make([]string, len(filters.SecurityTypes))
		for i, t := range filters.SecurityTypes {
			types[i] = string(t)
		}
		qb = qb.Where(sq.Expr("type = ANY(?)", types))
	}
	if filters.ISIN != "" {
		qb = qb.Where(sq.ILike{"isin": "%" + filters.ISIN + "%"})
	}
	if filters.Name != "" {
		qb = qb.Where(sq.ILike{"name": "%" + filters.Name + "%"})
	}
	if filters.Query != "" {
		q := "%" + filters.Query + "%"
		qb = qb.Where(sq.Or{
			sq.ILike{"name": q},
			sq.ILike{"isin": q},
		})
	}
	if !filters.TimeFilters.IssueAfter.IsZero() {
		qb = qb.Where(sq.GtOrEq{"issue_date": filters.TimeFilters.IssueAfter})
	}
	if !filters.TimeFilters.IssueBefore.IsZero() {
		qb = qb.Where(sq.LtOrEq{"issue_date": filters.TimeFilters.IssueBefore})
	}
	if !filters.TimeFilters.MaturityAfter.IsZero() {
		qb = qb.Where(sq.GtOrEq{"maturity_date": filters.TimeFilters.MaturityAfter})
	}
	if !filters.TimeFilters.MaturityBefore.IsZero() {
		qb = qb.Where(sq.LtOrEq{"maturity_date": filters.TimeFilters.MaturityBefore})
	}

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := repo.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list securities: %w", err)
	}
	defer rows.Close()

	var securities []sqlc.Security
	for rows.Next() {
		var sec sqlc.Security
		if err := rows.Scan(
			&sec.ID,
			&sec.PublicID,
			&sec.Isin,
			&sec.Name,
			&sec.Type,
			&sec.Serie,
			&sec.IssueDate,
			&sec.MaturityDate,
			&sec.NominalValue,
			&sec.Currency,
			&sec.Fx,
			&sec.Enable,
			&sec.CreatedAt,
			&sec.UpdatedAt,
		); err != nil {
			return nil, err
		}
		securities = append(securities, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*domain.Security, 0, len(securities))
	for _, sec := range securities {
		d, err := toDomain(sec)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}

	return result, nil
}

func (repo *SecurityRepo) GetEnableList(ctx context.Context) ([]*domain.Security, error) {
	securities, err := repo.queries.GetEnableList(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Security, len(securities))
	for i, sec := range securities {
		d, err := toDomain(sec)
		if err != nil {
			return nil, err
		}
		result[i] = d
	}

	return result, nil
}

func toDomain(sec sqlc.Security) (*domain.Security, error) {
	nominalVal, err := domain.NumericToFloat(sec.NominalValue)
	if err != nil {
		return nil, err
	}

	fxVal, err := domain.NumericToFloat(sec.Fx)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructSecurity(
		sec.ID,
		sec.PublicID,
		sec.Isin,
		sec.Name,
		sec.Type,
		sec.Enable.Bool,
		sec.Serie,
		sec.IssueDate.Time,
		sec.MaturityDate.Time,
		nominalVal,
		fxVal,
		sec.Currency,
		sec.CreatedAt.Time,
		sec.UpdatedAt.Time,
	)
}
