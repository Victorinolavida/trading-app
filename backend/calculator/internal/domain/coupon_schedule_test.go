package domain

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCouponSchedule(t *testing.T) {
	maturity := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	nominalValue, err := NewDecimal(int64(1000) * DecimalScale)
	require.NoError(t, err)

	schedule, err := BuildCouponSchedule(maturity, startDate, 182, 0.05, nominalValue)
	require.NoError(t, err)
	require.NotEmpty(t, schedule.CashFlows)

	last := schedule.CashFlows[len(schedule.CashFlows)-1]
	assert.True(t, last.Date.Equal(maturity))
	expectedFinal, err := NewDecimal(int64(math.Round(1.05 * float64(1000) * float64(DecimalScale))))
	require.NoError(t, err)
	assert.Equal(t, expectedFinal, last.Amount)
}

func TestBuildCouponSchedule_InvalidPeriod(t *testing.T) {
	_, err := BuildCouponSchedule(time.Now(), time.Now(), 0, 0.05, Decimal(0))
	assert.Error(t, err)
}
