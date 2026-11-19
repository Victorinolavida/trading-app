// Package api is the public boundary of the calculator bounded context.
package api

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/calculator/internal/application/ports"
	"github.com/victorinolavida/calculator/internal/domain"
	"github.com/victorinolavida/pkg/lifecycle"
)

// PriceRequest is the input of the price use case.
type PriceRequest struct {
	SecurityID     string
	SettlementDate time.Time
	Yield          float64
}

// PriceResult is the output of the price use case.
type PriceResult struct {
	CleanPrice      domain.Decimal
	AccruedInterest domain.Decimal
	DirtyPrice      domain.Decimal
	MXNValue        domain.Decimal
}

// YieldRequest is the input of the yield use case.
type YieldRequest struct {
	SecurityID     string
	SettlementDate time.Time
	CleanPrice     domain.Decimal
}

// CashflowRequest is the input of the cashflow use case.
type CashflowRequest struct {
	SecurityID     string
	SettlementDate time.Time
}

// Deps are the external dependencies the calculator module needs to start.
type Deps struct {
	Security ports.SecurityProvider
	Udi      ports.UdiProvider
}

// Module is the in-process calculator module.
type Module struct {
	security ports.SecurityProvider
	udi      ports.UdiProvider
}

var _ lifecycle.Service = (*Module)(nil)

// New wires the calculator module on top of the securities provider.
func New(deps Deps) (*Module, error) {
	return &Module{security: deps.Security, udi: deps.Udi}, nil
}

// Start runs the calculator module until the context is cancelled.
func (m *Module) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Close releases the calculator module.
func (m *Module) Close(context.Context) error {
	return nil
}

// Price computes clean price, accrued interest and dirty price for a security.
func (m *Module) Price(ctx context.Context, req PriceRequest) (PriceResult, error) {
	id, err := uuid.Parse(req.SecurityID)
	if err != nil {
		return PriceResult{}, err
	}
	security, err := m.security.GetByPublicID(ctx, id)
	if err != nil {
		return PriceResult{}, err
	}
	schedule, err := domain.BuildCouponSchedule(security.MaturityDate, req.SettlementDate, security.CouponPeriodDays, security.CouponRate, security.NominalValue)
	if err != nil {
		return PriceResult{}, err
	}
	daysSinceLastCoupon := daysSinceLastCoupon(req.SettlementDate, schedule)
	accrued := domain.ComputeAccruedInterest(security.NominalValue, security.CouponRate, daysSinceLastCoupon, security.DayCountBasis)
	dirty := domain.ComputeDirtyPrice(schedule, req.SettlementDate, req.Yield)
	clean := domain.ComputeCleanPrice(schedule, req.SettlementDate, req.Yield, security.NominalValue, security.CouponRate, daysSinceLastCoupon, security.DayCountBasis)
	mxn := clean
	if isUdibono(security.SecurityType) {
		udiValue, err := m.udi.GetByDate(ctx, req.SettlementDate)
		if err != nil {
			return PriceResult{}, err
		}
		mxn = domain.ComputeMXNValue(udiValue, clean)
	}
	return PriceResult{CleanPrice: clean, AccruedInterest: accrued, DirtyPrice: dirty, MXNValue: mxn}, nil
}

// Yield solves the yield from a clean price.
func (m *Module) Yield(ctx context.Context, req YieldRequest) (float64, error) {
	id, err := uuid.Parse(req.SecurityID)
	if err != nil {
		return 0, err
	}
	security, err := m.security.GetByPublicID(ctx, id)
	if err != nil {
		return 0, err
	}
	schedule, err := domain.BuildCouponSchedule(security.MaturityDate, req.SettlementDate, security.CouponPeriodDays, security.CouponRate, security.NominalValue)
	if err != nil {
		return 0, err
	}
	daysSinceLastCoupon := daysSinceLastCoupon(req.SettlementDate, schedule)
	return domain.ComputeYield(schedule, req.SettlementDate, req.CleanPrice, security.NominalValue, security.CouponRate, daysSinceLastCoupon, security.DayCountBasis), nil
}

// Cashflows returns the coupon schedule.
func (m *Module) Cashflows(ctx context.Context, req CashflowRequest) ([]domain.CashFlow, error) {
	id, err := uuid.Parse(req.SecurityID)
	if err != nil {
		return nil, err
	}
	security, err := m.security.GetByPublicID(ctx, id)
	if err != nil {
		return nil, err
	}
	schedule, err := domain.BuildCouponSchedule(security.MaturityDate, req.SettlementDate, security.CouponPeriodDays, security.CouponRate, security.NominalValue)
	if err != nil {
		return nil, err
	}
	return schedule.CashFlows, nil
}

// isUdibono reports whether the security is denominated in UDI.
func isUdibono(securityType string) bool {
	switch strings.ToUpper(strings.TrimSpace(securityType)) {
	case "UDI", "UDIBONO", "UD":
		return true
	default:
		return false
	}
}

// daysSinceLastCoupon returns the number of days elapsed since the last coupon
// before the settlement date.
func daysSinceLastCoupon(settlement time.Time, schedule *domain.CouponSchedule) int {
	last := settlement
	for _, cf := range schedule.CashFlows {
		if !cf.Date.After(settlement) && cf.Date.Before(last) {
			last = cf.Date
		}
	}
	days := int(settlement.Sub(last).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return days
}
