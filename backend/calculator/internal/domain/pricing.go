package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

const calculationPrecision = 24

var (
	one            = decimal.NewFromInt(1)
	two            = decimal.NewFromInt(2)
	yieldTolerance = decimal.New(1, -12)
	priceTolerance = decimal.New(1, -9)
)

func (s *CouponSchedule) periodFraction() decimal.Decimal {
	return decimal.NewFromInt(int64(s.CouponPeriodDays)).DivRound(decimal.NewFromInt(int64(s.DayCountBasis)), calculationPrecision)
}

// dirtyPrice returns the present value and its derivative with respect to
// annual yield. Only the first discount exponent can be fractional.
func (s *CouponSchedule) dirtyPrice(settlement time.Time, yield decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	period := s.periodFraction()
	base := one.Add(yield.Mul(period))
	if !base.IsPositive() {
		return decimal.Zero, decimal.Zero, errors.New("yield implies a nonpositive discount factor")
	}
	firstDays := daysBetween(CalendarDate(settlement), s.CashFlows[0].Date)
	exponent := decimal.NewFromInt(firstDays).DivRound(decimal.NewFromInt(int64(s.CouponPeriodDays)), calculationPrecision)
	factor, err := base.PowWithPrecision(exponent.Neg(), calculationPrecision)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	total, weightedTotal := decimal.Zero, decimal.Zero
	for i := range s.CashFlows {
		payment := s.coupon
		if i == len(s.CashFlows)-1 {
			payment = payment.Add(s.nominal)
		}
		presentValue := payment.Mul(factor).Round(calculationPrecision)
		total = total.Add(presentValue)
		weightedTotal = weightedTotal.Add(presentValue.Mul(exponent))
		exponent = exponent.Add(one)
		factor = factor.DivRound(base, calculationPrecision)
	}
	return total, weightedTotal.Neg().Mul(period).DivRound(base, calculationPrecision), nil
}

func (s *CouponSchedule) DirtyPrice(settlement time.Time, yield decimal.Decimal) (decimal.Decimal, error) {
	dirty, _, err := s.dirtyPrice(settlement, yield)
	return dirty, err
}

// Yield solves for the annual yield using Newton-Raphson within a bracket.
// When Newton would leave the bracket, the next step bisects it.
func (s *CouponSchedule) Yield(settlement time.Time, cleanPrice decimal.Decimal) (decimal.Decimal, error) {
	if !cleanPrice.IsPositive() {
		return decimal.Zero, errors.New("clean price must be positive")
	}
	if !s.CashFlows[len(s.CashFlows)-1].Date.After(CalendarDate(settlement)) {
		return decimal.Zero, errors.New("yield is undefined at maturity")
	}
	target := cleanPrice.Add(s.AccruedInterest)
	evaluate := func(yield decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
		price, slope, err := s.dirtyPrice(settlement, yield)
		return price.Sub(target), slope, err
	}

	atZero, _, err := evaluate(decimal.Zero)
	if err != nil || atZero.Abs().LessThanOrEqual(priceTolerance) {
		return decimal.Zero, err
	}

	low, high := decimal.Zero, decimal.Zero
	step := decimal.New(5, -2)
	floor := one.Neg().DivRound(s.periodFraction(), calculationPrecision)
	bracketed := false
	for range 64 {
		var candidate decimal.Decimal
		if atZero.IsPositive() {
			high = high.Add(step)
			step = step.Mul(two)
			candidate = high
		} else {
			low = low.Add(floor).DivRound(two, calculationPrecision)
			candidate = low
		}
		value, _, err := evaluate(candidate)
		if err != nil {
			return decimal.Zero, err
		}
		if value.Sign() != atZero.Sign() {
			bracketed = true
			break
		}
	}
	if !bracketed {
		return decimal.Zero, errors.New("could not bracket yield for the clean price")
	}

	yield := low.Add(high).DivRound(two, calculationPrecision)
	for range 100 {
		value, slope, err := evaluate(yield)
		if err != nil {
			return decimal.Zero, err
		}
		if value.Abs().LessThanOrEqual(priceTolerance) || high.Sub(low).LessThanOrEqual(yieldTolerance) {
			return yield, nil
		}
		if value.IsPositive() {
			low = yield
		} else {
			high = yield
		}
		if !slope.IsZero() {
			next := yield.Sub(value.DivRound(slope, calculationPrecision))
			if next.GreaterThan(low) && next.LessThan(high) {
				yield = next
				continue
			}
		}
		yield = low.Add(high).DivRound(two, calculationPrecision)
	}
	return decimal.Zero, errors.New("yield calculation did not converge")
}
