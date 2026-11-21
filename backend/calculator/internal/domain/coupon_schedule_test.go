package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCouponScheduleAndPricing(t *testing.T) {
	maturity := time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	settlement := maturity.AddDate(0, 0, -273)
	schedule, err := BuildCouponSchedule(maturity, settlement, 182, 360, decimal.RequireFromString("0.06"), Decimal(100*DecimalScale))
	require.NoError(t, err)
	require.Len(t, schedule.CashFlows, 2)
	require.Equal(t, maturity.AddDate(0, 0, -182), schedule.CashFlows[0].Date)
	require.Equal(t, maturity, schedule.CashFlows[1].Date)
	require.Equal(t, Decimal(3_033_333), schedule.CashFlows[0].Amount)
	require.Equal(t, Decimal(103_033_333), schedule.CashFlows[1].Amount)
	require.Equal(t, 91, schedule.DaysAccrued)
	require.Equal(t, "1.516667", schedule.AccruedInterest.Round(6).String())

	for _, rate := range []string{"-0.02", "0", "0.04", "0.25"} {
		yield := decimal.RequireFromString(rate)
		dirty, err := schedule.DirtyPrice(settlement, yield)
		require.NoError(t, err)
		clean := dirty.Sub(schedule.AccruedInterest)
		calculatedYield, err := schedule.Yield(settlement, clean)
		require.NoError(t, err)
		require.InDelta(t, yield.InexactFloat64(), calculatedYield.InexactFloat64(), 1e-9)
	}
	dirty, err := schedule.DirtyPrice(settlement, decimal.RequireFromString("0.04"))
	require.NoError(t, err)
	// The two cash flows are one half and one and a half coupon periods away.
	require.InDelta(t, 102.9882887021, dirty.InexactFloat64(), 1e-6)
}

func TestCouponScheduleOnPaymentDates(t *testing.T) {
	maturity := time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	for _, days := range []int{-182, 0} {
		settlement := maturity.AddDate(0, 0, days)
		schedule, err := BuildCouponSchedule(maturity, settlement, 182, 360, decimal.RequireFromString("0.06"), Decimal(100*DecimalScale))
		require.NoError(t, err)
		require.Zero(t, schedule.DaysAccrued)
		require.Len(t, schedule.CashFlows, 1)
		if days == 0 {
			price, err := schedule.DirtyPrice(settlement, decimal.RequireFromString("0.04"))
			require.NoError(t, err)
			require.Equal(t, "100", price.String())
		}
	}
}

func TestBuildCouponScheduleInvalidPeriod(t *testing.T) {
	_, err := BuildCouponSchedule(time.Now(), time.Now(), 0, 360, decimal.Zero, Decimal(100*DecimalScale))
	require.Error(t, err)
}
