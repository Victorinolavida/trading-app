package domain

import "math"

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
