package domain

import (
	"errors"
	"math"
	"time"
)

// CashFlow is a single cash flow of a fixed income security.
type CashFlow struct {
	Date   time.Time
	Amount Decimal
}

// CouponSchedule is the cash flow schedule of a fixed income security.
type CouponSchedule struct {
	MaturityDate     time.Time
	CouponPeriodDays int
	CouponRate       float64
	NominalValue     Decimal
	CashFlows        []CashFlow
}

// BuildCouponSchedule builds the cash flow schedule backward from maturity in
// couponPeriodDays steps, collecting cash flows until the start date. The final
// cash flow pays the nominal value and the preceding cash flows pay the periodic
// coupon.
func BuildCouponSchedule(
	maturity time.Time,
	startDate time.Time,
	couponPeriodDays int,
	couponRate float64,
	nominalValue Decimal,
) (*CouponSchedule, error) {
	if couponPeriodDays <= 0 {
		return nil, errors.New("coupon period days must be positive")
	}

	schedule := &CouponSchedule{
		MaturityDate:     maturity,
		CouponPeriodDays: couponPeriodDays,
		CouponRate:       couponRate,
		NominalValue:     nominalValue,
	}

	finalAmount, err := NewDecimal(int64(math.Round(float64(nominalValue.Value()) * (1 + couponRate))))
	if err != nil {
		return nil, err
	}
	final := CashFlow{Date: maturity, Amount: finalAmount}
	schedule.CashFlows = append(schedule.CashFlows, final)

	end := maturity.Add(-time.Duration(couponPeriodDays) * 24 * time.Hour)
	for !end.Before(startDate) {
		couponAmount, err := NewDecimal(int64(math.Round(float64(nominalValue.Value()) * couponRate)))
		if err != nil {
			return nil, err
		}
		schedule.CashFlows = append(schedule.CashFlows, CashFlow{Date: end, Amount: couponAmount})
		end = end.Add(-time.Duration(couponPeriodDays) * 24 * time.Hour)
	}

	for i, j := 0, len(schedule.CashFlows)-1; i < j; i, j = i+1, j-1 {
		schedule.CashFlows[i], schedule.CashFlows[j] = schedule.CashFlows[j], schedule.CashFlows[i]
	}

	return schedule, nil
}
