package security

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"time"
)

type UDIProvider struct{ Pool *pgxpool.Pool }

func (p UDIProvider) GetByDate(ctx context.Context, date time.Time) (decimal.Decimal, error) {
	var value string
	err := p.Pool.QueryRow(ctx, "SELECT value::text FROM udi_values WHERE date = $1", date).Scan(&value)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromString(value)
}
