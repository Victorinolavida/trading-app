package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/victorinolavida/calculator/internal/domain"
)

// Security holds the coupon data the calculator needs from the securities service.
type Security struct {
	ID               string
	PublicID         string
	ISIN             string
	Name             string
	SecurityType     string
	MaturityDate     time.Time
	NominalValue     domain.Decimal
	Currency         string
	CouponRate       decimal.Decimal
	CouponPeriodDays int
	DayCountBasis    int
}

// SecurityProvider fetches security details from the securities service.
type SecurityProvider interface {
	GetByPublicID(ctx context.Context, publicID uuid.UUID) (Security, error)
}
