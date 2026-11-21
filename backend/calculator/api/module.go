// Package api is the public boundary of the calculator bounded context.
package api

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/google/uuid"
	"github.com/victorinolavida/calculator/internal/application/ports"
	"github.com/victorinolavida/calculator/internal/domain"
	"github.com/victorinolavida/pkg/lifecycle"
)

// PriceRequest is the input of the price use case.
type PriceRequest struct {
	SecurityID     string    `json:"security_id" validate:"required"`
	SettlementDate time.Time `json:"settlement_date" validate:"required"`
	Yield          string    `json:"yield" validate:"required"`
}

// PriceResult is the output of the price use case.
type PriceResult struct {
	CleanPrice      string `json:"clean_price"`
	AccruedInterest string `json:"accrued_interest"`
	DirtyPrice      string `json:"dirty_price"`
	MXNValue        string `json:"mxn_value"`
}

// YieldRequest is the input of the yield use case.
type YieldRequest struct {
	SecurityID     string    `json:"security_id" validate:"required"`
	SettlementDate time.Time `json:"settlement_date" validate:"required"`
	CleanPrice     string    `json:"clean_price" validate:"required"`
}

// CashflowRequest is the input of the cashflow use case.
type CashflowRequest struct {
	SecurityID     string    `json:"security_id" validate:"required"`
	SettlementDate time.Time `json:"settlement_date"`
}

// Security is the public alias for the security data the calculator needs.
type Security = ports.Security

// SecurityProvider is the public alias for the security provider interface.
type SecurityProvider = ports.SecurityProvider

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

// Price computes bond prices and converts Udibono clean prices into MXN.
func (m *Module) Price(ctx context.Context, req PriceRequest) (PriceResult, error) {
	security, clean, dirty, accrued, err := m.calculatePrices(ctx, req)
	if err != nil {
		return PriceResult{}, err
	}
	mxn := clean
	if isUdibono(security) {
		if m.udi == nil {
			return PriceResult{}, errors.New("UDI quote provider is unavailable")
		}
		quote, err := m.udi.GetByDate(ctx, domain.CalendarDate(req.SettlementDate))
		if err != nil {
			return PriceResult{}, err
		}
		if !quote.IsPositive() {
			return PriceResult{}, errors.New("UDI quote must be positive")
		}
		mxn = clean.Mul(quote)
	}
	return PriceResult{
		CleanPrice: clean.Round(6).String(), AccruedInterest: accrued.Round(6).String(),
		DirtyPrice: dirty.Round(6).String(), MXNValue: mxn.Round(6).String(),
	}, nil
}

// CleanPrice computes the order price in the security's native currency.
func (m *Module) CleanPrice(ctx context.Context, req PriceRequest) (string, error) {
	_, clean, _, _, err := m.calculatePrices(ctx, req)
	if err != nil {
		return "", err
	}
	return clean.Round(6).String(), nil
}

func (m *Module) calculatePrices(ctx context.Context, req PriceRequest) (Security, decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	yield, err := decimal.NewFromString(req.Yield)
	if err != nil {
		return Security{}, decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	security, schedule, err := m.schedule(ctx, req.SecurityID, req.SettlementDate)
	if err != nil {
		return Security{}, decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	dirty, err := schedule.DirtyPrice(req.SettlementDate, yield)
	return security, dirty.Sub(schedule.AccruedInterest), dirty, schedule.AccruedInterest, err
}

// Yield solves the annual yield from a clean price for the numeric HTTP contract.
func (m *Module) Yield(ctx context.Context, req YieldRequest) (float64, error) {
	yield, err := m.YieldDecimal(ctx, req)
	if err != nil {
		return 0, err
	}
	return yield.InexactFloat64(), nil
}

// YieldDecimal solves the annual yield without converting through a float.
func (m *Module) YieldDecimal(ctx context.Context, req YieldRequest) (decimal.Decimal, error) {
	cleanPrice, err := decimal.NewFromString(req.CleanPrice)
	if err != nil {
		return decimal.Zero, err
	}
	_, schedule, err := m.schedule(ctx, req.SecurityID, req.SettlementDate)
	if err != nil {
		return decimal.Zero, err
	}
	return schedule.Yield(req.SettlementDate, cleanPrice)
}

// Cashflows returns the coupon schedule.
func (m *Module) Cashflows(ctx context.Context, req CashflowRequest) ([]domain.CashFlow, error) {
	settlement := req.SettlementDate
	if settlement.IsZero() {
		settlement = time.Now().UTC()
	}
	_, schedule, err := m.schedule(ctx, req.SecurityID, settlement)
	if err != nil {
		return nil, err
	}
	return schedule.CashFlows, nil
}

func (m *Module) schedule(ctx context.Context, securityID string, settlement time.Time) (Security, *domain.CouponSchedule, error) {
	id, err := uuid.Parse(securityID)
	if err != nil {
		return Security{}, nil, err
	}
	security, err := m.security.GetByPublicID(ctx, id)
	if err != nil {
		return Security{}, nil, err
	}
	schedule, err := domain.BuildCouponSchedule(security.MaturityDate, settlement, security.CouponPeriodDays, security.DayCountBasis, security.CouponRate, security.NominalValue)
	return security, schedule, err
}

// ParseDecimal converts a decimal string into the calculator's scaled integer.
func ParseDecimal(value string) (domain.Decimal, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return 0, err
	}
	scaled := d.Shift(6).Truncate(0)
	if scaled.IsNegative() || scaled.GreaterThan(decimal.NewFromInt(domain.MaxDecimalValue)) {
		return 0, domain.ErrInvalidNumericNumber
	}
	return domain.Decimal(scaled.IntPart()), nil
}

func isUdibono(security Security) bool {
	if strings.EqualFold(strings.TrimSpace(security.Currency), "UDI") {
		return true
	}
	switch strings.ToUpper(strings.TrimSpace(security.SecurityType)) {
	case "UDI", "UDIBONO", "UD":
		return true
	default:
		return false
	}
}

// NewBondSecurity adapts security reference data to the calculator's bond inputs.
func NewBondSecurity(id, isin, name, securityType, currency string, maturityDate time.Time, nominal domain.Decimal) Security {
	return Security{
		ID: id, PublicID: id, ISIN: isin, Name: name, SecurityType: securityType, Currency: currency,
		MaturityDate: maturityDate, NominalValue: nominal, CouponRate: couponRateFromName(name),
		CouponPeriodDays: 182, DayCountBasis: 360,
	}
}

func couponRateFromName(name string) decimal.Decimal {
	for _, field := range strings.Fields(name) {
		if rate, err := decimal.NewFromString(strings.TrimSuffix(field, "%")); err == nil && !rate.IsNegative() {
			return rate.Shift(-2)
		}
	}
	return decimal.New(5, -2)
}
