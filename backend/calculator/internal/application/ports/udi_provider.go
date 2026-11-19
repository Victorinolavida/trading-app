package ports

import (
	"context"
	"time"

	"github.com/victorinolavida/calculator/internal/domain"
)

// UdiProvider fetches UDI values from the securities service.
type UdiProvider interface {
	GetByDate(ctx context.Context, date time.Time) (domain.Decimal, error)
}
