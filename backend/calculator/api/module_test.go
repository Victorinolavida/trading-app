package api

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

type securityStub struct{ security Security }

func (s securityStub) GetByPublicID(context.Context, uuid.UUID) (Security, error) {
	return s.security, nil
}

type udiStub func(time.Time) (decimal.Decimal, error)

func (lookup udiStub) GetByDate(_ context.Context, date time.Time) (decimal.Decimal, error) {
	return lookup(date)
}

func TestPriceConvertsUDISecurityBySettlementDate(t *testing.T) {
	id := uuid.New()
	settlement := time.Date(2027, 1, 5, 0, 0, 0, 0, time.UTC)
	nominal, err := ParseDecimal("100")
	require.NoError(t, err)
	security := NewBondSecurity(id.String(), "", "MUDI 4.5", "S", "UDI", settlement.AddDate(0, 0, 91), nominal)
	module, err := New(Deps{
		Security: securityStub{security},
		Udi: udiStub(func(date time.Time) (decimal.Decimal, error) {
			require.Equal(t, settlement, date)
			return decimal.RequireFromString("8.123456"), nil
		}),
	})
	require.NoError(t, err)

	result, err := module.Price(t.Context(), PriceRequest{id.String(), settlement, "0.04"})
	require.NoError(t, err)
	clean := decimal.RequireFromString(result.CleanPrice)
	mxn := decimal.RequireFromString(result.MXNValue)
	require.InDelta(t, clean.Mul(decimal.RequireFromString("8.123456")).InexactFloat64(), mxn.InexactFloat64(), 0.000005)

	module.udi = nil
	_, err = module.Price(t.Context(), PriceRequest{id.String(), settlement, "0.04"})
	require.ErrorContains(t, err, "UDI quote provider")
	native, err := module.CleanPrice(t.Context(), PriceRequest{id.String(), settlement, "0.04"})
	require.NoError(t, err)
	require.Equal(t, result.CleanPrice, native)
}
