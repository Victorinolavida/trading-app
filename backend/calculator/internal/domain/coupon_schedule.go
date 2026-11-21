package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// CashFlow is a single cash flow of a fixed income security.
type CashFlow struct {
	Date   time.Time
	Amount Decimal
}

type CouponSchedule struct {
	CouponPeriodDays int
	DayCountBasis    int
	DaysAccrued      int
	AccruedInterest  decimal.Decimal
	CashFlows        []CashFlow
	coupon           decimal.Decimal
	nominal          decimal.Decimal
}

func CalendarDate(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysBetween(start, end time.Time) int64 {
	return (end.Unix() - start.Unix()) / 86_400
}

// BuildCouponSchedule walks backward from maturity. Payments on settlement
// belong to the previous holder, except for principal at maturity.
func BuildCouponSchedule(
	maturity, settlement time.Time,
	periodDays, dayCountBasis int,
	couponRate decimal.Decimal,
	nominalValue Decimal,
) (*CouponSchedule, error) {
	if periodDays <= 0 || dayCountBasis <= 0 || couponRate.IsNegative() || nominalValue <= 0 {
		return nil, errors.New("invalid coupon schedule inputs")
	}
	if maturity.IsZero() || settlement.IsZero() {
		return nil, errors.New("maturity and settlement dates are required")
	}
	maturity, settlement = CalendarDate(maturity), CalendarDate(settlement)
	if maturity.Before(settlement) {
		return nil, errors.New("settlement date is after maturity")
	}

	nominal := decimal.New(int64(nominalValue), -6)
	annualCoupon := nominal.Mul(couponRate)
	basis := decimal.NewFromInt(int64(dayCountBasis))
	coupon := annualCoupon.Mul(decimal.NewFromInt(int64(periodDays))).DivRound(basis, calculationPrecision)
	schedule := &CouponSchedule{
		CouponPeriodDays: periodDays,
		DayCountBasis:    dayCountBasis,
		coupon:           coupon,
		nominal:          nominal,
	}
	if maturity.Equal(settlement) {
		schedule.CashFlows = []CashFlow{{Date: maturity, Amount: nominalValue}}
		schedule.coupon = decimal.Zero
		return schedule, nil
	}

	lastCoupon := maturity
	for lastCoupon.After(settlement) {
		amount := coupon
		if lastCoupon.Equal(maturity) {
			amount = amount.Add(nominal)
		}
		scaled := amount.Shift(6).Round(0)
		if scaled.GreaterThan(decimal.NewFromInt(MaxDecimalValue)) {
			return nil, ErrInvalidNumericNumber
		}
		schedule.CashFlows = append(schedule.CashFlows, CashFlow{Date: lastCoupon, Amount: Decimal(scaled.IntPart())})
		lastCoupon = lastCoupon.AddDate(0, 0, -periodDays)
	}
	schedule.DaysAccrued = int(daysBetween(lastCoupon, settlement))
	schedule.AccruedInterest = annualCoupon.Mul(decimal.NewFromInt(int64(schedule.DaysAccrued))).DivRound(basis, calculationPrecision)

	for i, j := 0, len(schedule.CashFlows)-1; i < j; i, j = i+1, j-1 {
		schedule.CashFlows[i], schedule.CashFlows[j] = schedule.CashFlows[j], schedule.CashFlows[i]
	}
	return schedule, nil
}
