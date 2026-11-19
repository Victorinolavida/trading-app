package domain

import (
	"math"
	"time"
)

// ComputeAccruedInterest returns the accrued interest for the period ending at
// settlement. It is computed as nominal value * coupon rate * days since the last
// coupon divided by the day count basis.
func ComputeAccruedInterest(nominalValue Decimal, couponRate float64, daysSinceLastCoupon int, dayCountBasis int) Decimal {
	if dayCountBasis <= 0 {
		return Decimal(0)
	}
	amount, err := NewDecimal(int64(math.Round(float64(nominalValue.Value()) * couponRate * float64(daysSinceLastCoupon) / float64(dayCountBasis))))
	if err != nil {
		return Decimal(0)
	}
	return amount
}

// ComputeDirtyPrice discounts the coupon schedule cash flows at the given
// yield. The first cash flow may use a fractional exponent because it can occur
// before a full coupon period.
func ComputeDirtyPrice(schedule *CouponSchedule, settlementDate time.Time, yield float64) Decimal {
	discountRate := 1 + yield
	total := 0.0
	for _, cf := range schedule.CashFlows {
		days := int(cf.Date.Sub(settlementDate).Hours() / 24)
		if days < 0 {
			days = 0
		}
		exponent := float64(days) / float64(schedule.CouponPeriodDays)
		discounted := float64(cf.Amount.Value()) / math.Pow(discountRate, exponent)
		total += discounted
	}
	amount, err := NewDecimal(int64(math.Round(total)))
	if err != nil {
		return Decimal(0)
	}
	return amount
}

// ComputeCleanPrice returns the clean price as the dirty price minus the accrued
// interest.
func ComputeCleanPrice(
	schedule *CouponSchedule,
	settlementDate time.Time,
	yield float64,
	nominalValue Decimal,
	couponRate float64,
	daysSinceLastCoupon int,
	dayCountBasis int,
) Decimal {
	dirty := ComputeDirtyPrice(schedule, settlementDate, yield)
	accrued := ComputeAccruedInterest(nominalValue, couponRate, daysSinceLastCoupon, dayCountBasis)
	amount, err := NewDecimal(int64(math.Round(float64(dirty.Value()) - float64(accrued.Value()))))
	if err != nil {
		return Decimal(0)
	}
	return amount
}

// ComputeMXNValue converts a UDI-denominated result into MXN by multiplying it
// by the UDI value on the settlement date.
func ComputeMXNValue(udiValue Decimal, result Decimal) Decimal {
	resultUnits := float64(result.Value()) / float64(DecimalScale)
	udiUnits := float64(udiValue.Value()) / float64(DecimalScale)
	mxn := resultUnits * udiUnits
	amount, err := NewDecimal(int64(math.Round(mxn * float64(DecimalScale))))
	if err != nil {
		return Decimal(0)
	}
	return amount
}

// ComputeYield solves the yield from a clean price using Newton-Raphson with a
// bisection fallback.
func ComputeYield(
	schedule *CouponSchedule,
	settlementDate time.Time,
	cleanPrice Decimal,
	nominalValue Decimal,
	couponRate float64,
	daysSinceLastCoupon int,
	dayCountBasis int,
) float64 {
	target := float64(cleanPrice.Value()) / float64(DecimalScale)

	y := 0.05
	for i := 0; i < 100; i++ {
		clean := float64(ComputeCleanPrice(schedule, settlementDate, y, nominalValue, couponRate, daysSinceLastCoupon, dayCountBasis).Value()) / float64(DecimalScale)
		f := clean - target
		h := 1e-6
		cleanPrev := float64(ComputeCleanPrice(schedule, settlementDate, y-h, nominalValue, couponRate, daysSinceLastCoupon, dayCountBasis).Value()) / float64(DecimalScale)
		fp := (clean - cleanPrev) / h
		if fp == 0 {
			break
		}
		yNew := y - f/fp
		if yNew < 0 {
			yNew = 0
		}
		if math.Abs(yNew-y) < 1e-12 {
			y = yNew
			break
		}
		y = yNew
	}

	low, high := 0.0, 10.0
	for i := 0; i < 200; i++ {
		mid := (low + high) / 2
		clean := float64(ComputeCleanPrice(schedule, settlementDate, mid, nominalValue, couponRate, daysSinceLastCoupon, dayCountBasis).Value()) / float64(DecimalScale)
		if math.Abs(clean-target) < 1e-12 {
			return mid
		}
		if clean > target {
			low = mid
		} else {
			high = mid
		}
	}
	return (low + high) / 2
}
