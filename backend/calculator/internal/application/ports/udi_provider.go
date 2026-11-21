package ports

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// UdiProvider fetches UDI values from the securities service.
type UdiProvider interface {
	GetByDate(ctx context.Context, date time.Time) (decimal.Decimal, error)
}
