package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

const (
	testPrice    = 100_000
	testYield    = 5_000
	testNominal  = 1_000_000
	testFx       = 1_000_000
	testQuantity = 10
)

var expiration = time.Duration(1 * time.Hour)

func newTestOrder(t *testing.T, direction, orderType string, quantity int, price, yield int64) *domain.Order {
	t.Helper()
	order, err := domain.NewOrder(
		uuid.New(), uuid.New(),
		direction, orderType, quantity,
		price, yield,
		domain.Decimal(testNominal), domain.Decimal(testFx),
		expiration,
	)
	require.NoError(t, err)
	return order
}

func TestNewOrder_Limit(t *testing.T) {
	order := newTestOrder(t, "bid", "limit", testQuantity, testPrice, testYield)

	require.Equal(t, domain.DirectionBid, order.Direction())
	require.Equal(t, domain.OrderTypeLimit, order.OrderType())
	require.True(t, order.IsLimit())
	require.False(t, order.IsMarket())
	require.Equal(t, domain.Quantity(testQuantity), order.Quantity())
	require.Equal(t, domain.OrderStatusActive, order.Status())
	require.Equal(t, domain.AllocationStatusNone, order.AllocationStatus())
	require.Equal(t, domain.Quantity(0), order.AllocationQty())
	require.Equal(t, domain.Quantity(testQuantity), order.Remaining())

	require.NotNil(t, order.Price())
	require.Equal(t, domain.Decimal(testPrice), *order.Price())
	require.NotNil(t, order.Yield())
	require.Equal(t, domain.Decimal(testYield), *order.Yield())
	require.Equal(t, domain.Decimal(testNominal), order.NominalValue())
	require.Equal(t, domain.Decimal(testFx), order.Fx())
}

func TestNewOrder_Market_HasNoPriceOrYield(t *testing.T) {
	order := newTestOrder(t, "ask", "market", testQuantity, 0, 0)

	require.True(t, order.IsMarket())
	require.Nil(t, order.Price())
	require.Nil(t, order.Yield())
	require.Equal(t, domain.OrderStatusActive, order.Status())
}

func TestNewOrder_Validation(t *testing.T) {
	tests := []struct {
		name      string
		traderID  uuid.UUID
		security  uuid.UUID
		direction string
		orderType string
		quantity  int
		price     int64
		yield     int64
		nominal   domain.Decimal
		fx        domain.Decimal
		wantErr   error
	}{
		{"nil trader id", uuid.Nil, uuid.New(), "bid", "limit", 10, testPrice, 0, testNominal, testFx, domainerrors.ErrInvalidTraderID},
		{"nil security id", uuid.New(), uuid.Nil, "bid", "limit", 10, testPrice, 0, testNominal, testFx, domainerrors.ErrInvalidSecurityID},
		{"bad direction", uuid.New(), uuid.New(), "buy", "limit", 10, testPrice, 0, testNominal, testFx, domainerrors.ErrInvalidDirection},
		{"bad order type", uuid.New(), uuid.New(), "bid", "stop", 10, testPrice, 0, testNominal, testFx, domainerrors.ErrInvalidOrderType},
		{"zero quantity", uuid.New(), uuid.New(), "bid", "limit", 0, testPrice, 0, testNominal, testFx, domainerrors.ErrInvalidQuantity},
		{"zero nominal", uuid.New(), uuid.New(), "bid", "limit", 10, testPrice, 0, 0, testFx, domainerrors.ErrInvalidNominalValue},
		{"zero fx", uuid.New(), uuid.New(), "bid", "limit", 10, testPrice, 0, testNominal, 0, domainerrors.ErrInvalidFX},
		{"limit without price", uuid.New(), uuid.New(), "bid", "limit", 10, 0, 0, testNominal, testFx, domainerrors.ErrInvalidPrice},
		{"limit negative price", uuid.New(), uuid.New(), "bid", "limit", 10, -1, 0, testNominal, testFx, domainerrors.ErrInvalidPrice},
		{"limit negative yield", uuid.New(), uuid.New(), "bid", "limit", 10, testPrice, -1, testNominal, testFx, domainerrors.ErrInvalidYield},
		{"market with price", uuid.New(), uuid.New(), "bid", "market", 10, testPrice, 0, testNominal, testFx, domainerrors.ErrPriceInMarketOrder},
		{"market with yield", uuid.New(), uuid.New(), "bid", "market", 10, 0, testYield, testNominal, testFx, domainerrors.ErrYieldInMarketOrder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewOrder(tt.traderID, tt.security, tt.direction, tt.orderType, tt.quantity, tt.price, tt.yield, tt.nominal, tt.fx, expiration)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestOrder_CancelKillExpire(t *testing.T) {
	order := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)

	require.NoError(t, order.Cancel())
	require.Equal(t, domain.OrderStatusCancelled, order.Status())
	require.ErrorIs(t, order.Cancel(), domainerrors.ErrOrderAlreadyCanceled)

	order.Kill()
	require.Equal(t, domain.OrderStatusKilled, order.Status())

	order.Expire()
	require.Equal(t, domain.OrderStatusExpired, order.Status())
}

func TestMatchWith_MarketTakerGetsMakerPrice(t *testing.T) {
	maker := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)
	taker := newTestOrder(t, "ask", "market", testQuantity, 0, 0)

	// Market taker has no price of its own.
	require.Nil(t, taker.Price())

	execution, ok := taker.MatchWith(maker)
	require.True(t, ok)

	// Execution price must be the maker's (resting order's) price.
	require.Equal(t, *maker.Price(), execution.Price)
	require.Equal(t, domain.Decimal(testPrice), execution.Price)
	require.Equal(t, domain.DirectionAsk, execution.Taker)
	require.Equal(t, domain.Quantity(testQuantity), execution.Qty)

	require.Equal(t, domain.Quantity(0), taker.Remaining())
	require.Equal(t, domain.Quantity(0), maker.Remaining())
	require.Equal(t, domain.AllocationStatusAllocated, taker.AllocationStatus())
	require.Equal(t, domain.AllocationStatusAllocated, maker.AllocationStatus())
	require.Equal(t, domain.OrderStatusFullyAllocated, taker.Status())
	require.Equal(t, domain.OrderStatusFullyAllocated, maker.Status())
}

func TestMatchWith_LimitTakerUsesOwnPrice(t *testing.T) {
	maker := newTestOrder(t, "bid", "limit", testQuantity, 90_000, 0)
	taker := newTestOrder(t, "ask", "limit", testQuantity, 100_000, 0)

	execution, ok := taker.MatchWith(maker)
	require.True(t, ok)
	require.Equal(t, domain.Decimal(100_000), execution.Price)
}

func TestMatchWith_PartialAllocation(t *testing.T) {
	maker := newTestOrder(t, "bid", "limit", 100, testPrice, 0)
	taker := newTestOrder(t, "ask", "limit", testQuantity, testPrice, 0)

	execution, ok := taker.MatchWith(maker)
	require.True(t, ok)
	require.Equal(t, domain.Quantity(testQuantity), execution.Qty)
	require.Equal(t, domain.Decimal(testPrice), execution.Price)

	require.Equal(t, domain.OrderStatusFullyAllocated, taker.Status())
	require.Equal(t, domain.AllocationStatusPartial, maker.AllocationStatus())
	require.Equal(t, domain.Quantity(90), maker.Remaining())
	require.Equal(t, domain.OrderStatusActive, maker.Status())
}

func TestMatchWith_NoMatch(t *testing.T) {
	t.Run("same direction", func(t *testing.T) {
		a := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)
		b := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)

		_, ok := a.MatchWith(b)
		require.False(t, ok)
	})

	t.Run("cancelled maker", func(t *testing.T) {
		maker := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)
		require.NoError(t, maker.Cancel())
		taker := newTestOrder(t, "ask", "limit", testQuantity, testPrice, 0)

		_, ok := taker.MatchWith(maker)
		require.False(t, ok)
	})

	t.Run("fully allocated maker", func(t *testing.T) {
		maker := newTestOrder(t, "bid", "limit", testQuantity, testPrice, 0)
		taker := newTestOrder(t, "ask", "limit", testQuantity, testPrice, 0)

		_, ok := taker.MatchWith(maker)
		require.True(t, ok)

		another := newTestOrder(t, "ask", "limit", testQuantity, testPrice, 0)
		_, ok = another.MatchWith(maker)
		require.False(t, ok)
	})
}

func TestOrder_AllocationPtc(t *testing.T) {
	price := domain.Decimal(testPrice)
	yield := domain.Decimal(testYield)

	newOrder := func(t *testing.T, quantity, allocated domain.Quantity) *domain.Order {
		t.Helper()
		status := domain.AllocationStatusNone
		if allocated > 0 {
			status = domain.AllocationStatusPartial
		}
		if allocated == quantity {
			status = domain.AllocationStatusAllocated
		}
		order, err := domain.ReconstructOrder(
			uuid.New(), uuid.New(), uuid.New(), uuid.New(),
			"bid", "limit", "active", string(status),
			quantity, allocated,
			&price, &yield,
			domain.Decimal(testNominal), domain.Decimal(testFx),
			time.Now(), time.Now(),
			time.Now().Add(expiration),
		)
		require.NoError(t, err)
		return order
	}

	tests := []struct {
		name      string
		quantity  domain.Quantity
		allocated domain.Quantity
		want      domain.Decimal
		wantStr   string
	}{
		{"zero allocated", 10, 0, 0, "0"},
		{"half allocated", 10, 5, 50_000_000, "50"},
		{"fully allocated", 10, 10, 100_000_000, "100"},
		{"one third", 3, 1, 33_333_333, "33.333333"},
		{"one sixth", 6, 1, 16_666_666, "16.666666"},
		{"one eighth has fractional percent", 8, 1, 12_500_000, "12.5"},
		{"almost fully allocated", 1_000, 999, 99_900_000, "99.9"},
		{"large quantities", 1_000_000_000, 500_000_000, 50_000_000, "50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newOrder(t, tt.quantity, tt.allocated).AllocationPtc()
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantStr, got.String())
		})
	}
}

func TestReconstructOrder(t *testing.T) {
	price := domain.Decimal(testPrice)
	yield := domain.Decimal(testYield)

	order, err := domain.ReconstructOrder(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"bid", "limit", "active", "partial",
		domain.Quantity(10), domain.Quantity(4),
		&price, &yield,
		domain.Decimal(testNominal), domain.Decimal(testFx),
		time.Now(), time.Now(),
		time.Now().Add(expiration),
	)
	require.NoError(t, err)
	require.Equal(t, domain.DirectionBid, order.Direction())
	require.Equal(t, domain.AllocationStatusPartial, order.AllocationStatus())
	require.Equal(t, domain.Quantity(6), order.Remaining())
	require.Equal(t, &price, order.Price())

	_, err = domain.ReconstructOrder(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"bid", "limit", "nope", "none",
		domain.Quantity(10), domain.Quantity(0),
		&price, &yield,
		domain.Decimal(testNominal), domain.Decimal(testFx),
		time.Now(), time.Now(),
		time.Now().Add(expiration),
	)
	require.ErrorIs(t, err, domainerrors.ErrInvalidOrderStatus)

	_, err = domain.ReconstructOrder(
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		"bid", "limit", "active", "none",
		domain.Quantity(10), domain.Quantity(11),
		&price, &yield,
		domain.Decimal(testNominal), domain.Decimal(testFx),
		time.Now(), time.Now(),
		time.Now().Add(expiration),
	)
	require.ErrorIs(t, err, domain.ErrInvalidNumericNumber)
}

func Test_OrderLimitExpiration(t *testing.T) {
	exp := time.Duration(0)
	_, err := domain.NewOrder(
		uuid.New(), uuid.New(),
		"ask",
		"limit",
		10,
		100,
		10,
		domain.Decimal(testNominal), domain.Decimal(testFx),
		exp,
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrInvalidExpiration)
}
