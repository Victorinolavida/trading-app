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
