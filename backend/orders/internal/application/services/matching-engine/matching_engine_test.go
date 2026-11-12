package matchingengine

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorinolavida/orders/internal/application/mocks"
	"github.com/victorinolavida/orders/internal/domain"
)

var expiration = time.Duration(1 * time.Hour)

func Test_SingleMatchExact(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	// script matching for
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	maker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"ask",
		"limit",
		100,
		981,
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	taker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"bid",
		"limit",
		100, // same qty
		981, // same price
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)

	assert.Len(t, matches, 1)
	assert.Len(t, trades, 1)

	assert.Equal(t, domain.AllocationStatusAllocated.String(), taker.AllocationStatus().String())
	assert.Equal(t, domain.AllocationStatusAllocated.String(), maker.AllocationStatus().String())

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), taker.Status().String())
	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker.Status().String())

	assert.Equal(t, domain.Quantity(0), taker.Remaining())
	assert.Equal(t, domain.Quantity(0), maker.Remaining())

	assert.Equal(t, domain.Quantity(100), taker.AllocationQty())
	assert.Equal(t, domain.Quantity(100), maker.AllocationQty())

	trade := trades[0]
	assert.Equal(t, maker.ID(), trade.MakerOrderID)
	assert.Equal(t, taker.ID(), trade.TakerOrderID)
	assert.Equal(t, taker.Direction(), trade.TakerSide)
	assert.Equal(t, domain.Quantity(100), trade.Quantity)
	assert.Equal(t, domain.Decimal(981), trade.Price)

	repo.AssertExpectations(t)
}

func Test_SingleMatchTakerMoreQty(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	// script matching for
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)
	maker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"ask",
		"limit",
		100,
		981,
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	taker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"bid",
		"limit",
		150, // same qty
		981, // same price
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)

	assert.Len(t, matches, 1)
	assert.Len(t, trades, 1)

	assert.Equal(t, domain.AllocationStatusPartial.String(), taker.AllocationStatus().String())
	assert.Equal(t, domain.AllocationStatusAllocated.String(), maker.AllocationStatus().String())

	assert.Equal(t, domain.OrderStatusActive.String(), taker.Status().String())
	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker.Status().String())

	assert.Equal(t, domain.Quantity(50), taker.Remaining())
	assert.Equal(t, domain.Quantity(0), maker.Remaining())

	assert.Equal(t, domain.Quantity(100), taker.AllocationQty())
	assert.Equal(t, domain.Quantity(100), maker.AllocationQty())

	trade := trades[0]
	assert.Equal(t, maker.ID(), trade.MakerOrderID)
	assert.Equal(t, taker.ID(), trade.TakerOrderID)
	assert.Equal(t, taker.Direction(), trade.TakerSide)
	assert.Equal(t, domain.Quantity(100), trade.Quantity)
	assert.Equal(t, domain.Decimal(981), trade.Price)
	repo.AssertExpectations(t)
}

func Test_SingleMatchMakerMoreQty(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	// script matching for
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)
	maker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"ask",
		"limit",
		200,
		981,
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	taker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"bid",
		"limit",
		100, // same qty
		981, // same price
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)

	assert.Len(t, matches, 1)
	assert.Len(t, trades, 1)

	assert.Equal(t, domain.AllocationStatusAllocated.String(), taker.AllocationStatus().String())
	assert.Equal(t, domain.AllocationStatusPartial.String(), maker.AllocationStatus().String())

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), taker.Status().String())
	assert.Equal(t, domain.OrderStatusActive.String(), maker.Status().String())

	assert.Equal(t, domain.Quantity(0), taker.Remaining())
	assert.Equal(t, domain.Quantity(100), maker.Remaining())

	assert.Equal(t, domain.Quantity(100), taker.AllocationQty())
	assert.Equal(t, domain.Quantity(100), maker.AllocationQty())

	trade := trades[0]
	assert.Equal(t, maker.ID(), trade.MakerOrderID)
	assert.Equal(t, taker.ID(), trade.TakerOrderID)
	assert.Equal(t, taker.Direction(), trade.TakerSide)
	assert.Equal(t, domain.Quantity(100), trade.Quantity)
	assert.Equal(t, domain.Decimal(981), trade.Price)
	repo.AssertExpectations(t)
}

func Test_SingleMatchMarketTakerUsesMakerPrice(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	maker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"ask",
		"limit",
		100,
		981,
		100,
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	taker, err := domain.NewOrder(
		uuid.New(),
		uuid.New(),
		"bid",
		"market",
		100, // same qty
		0,   // market orders have no price
		0,   // and no yield
		domain.Decimal(100),
		domain.Decimal(1),
		expiration,
	)
	assert.NoError(t, err)
	assert.Nil(t, taker.Price())

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)

	assert.Len(t, matches, 1)
	assert.Len(t, trades, 1)

	assert.Equal(t, domain.AllocationStatusAllocated.String(), taker.AllocationStatus().String())
	assert.Equal(t, domain.AllocationStatusAllocated.String(), maker.AllocationStatus().String())

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), taker.Status().String())
	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker.Status().String())

	assert.Equal(t, domain.Quantity(0), taker.Remaining())
	assert.Equal(t, domain.Quantity(0), maker.Remaining())

	trade := trades[0]
	assert.Equal(t, maker.ID(), trade.MakerOrderID)
	assert.Equal(t, taker.ID(), trade.TakerOrderID)
	assert.Equal(t, taker.Direction(), trade.TakerSide)
	assert.Equal(t, domain.Quantity(100), trade.Quantity)
	// Market taker has no price, so the trade executes at the maker's price.
	assert.Equal(t, *maker.Price(), trade.Price)
	assert.Equal(t, domain.Decimal(981), trade.Price)
	repo.AssertExpectations(t)
}

func Test_MultipleMatchesExact(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	// GetAsksToTrade orders by price ASC, so the book comes back cheapest first.
	maker1, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "limit", 50, 981, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)
	maker2, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "limit", 50, 982, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker1, maker2}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	taker, err := domain.NewOrder(uuid.New(), uuid.New(), "bid", "market", 100, 0, 0, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Len(t, trades, 2)

	// Trades follow the book order: cheapest ask first.
	assert.Equal(t, maker1.ID(), trades[0].MakerOrderID)
	assert.Equal(t, domain.Decimal(981), trades[0].Price)
	assert.Equal(t, domain.Quantity(50), trades[0].Quantity)
	assert.Equal(t, maker2.ID(), trades[1].MakerOrderID)
	assert.Equal(t, domain.Decimal(982), trades[1].Price)
	assert.Equal(t, domain.Quantity(50), trades[1].Quantity)

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), taker.Status().String())
	assert.Equal(t, domain.Quantity(0), taker.Remaining())
	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker1.Status().String())
	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker2.Status().String())
	repo.AssertExpectations(t)
}

func Test_MultipleMatchesLastMakerHasMoreQty(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	// GetAsksToTrade orders by price ASC, so the book comes back cheapest first.
	maker1, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "limit", 50, 981, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)
	maker2, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "limit", 200, 982, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker1, maker2}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	// Bid limit price must cross both asks (981 and 982).
	taker, err := domain.NewOrder(uuid.New(), uuid.New(), "bid", "limit", 100, 982, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Len(t, trades, 2)

	assert.Equal(t, maker1.ID(), trades[0].MakerOrderID)
	assert.Equal(t, domain.Quantity(50), trades[0].Quantity)
	assert.Equal(t, domain.Decimal(982), trades[0].Price)
	assert.Equal(t, maker2.ID(), trades[1].MakerOrderID)
	assert.Equal(t, domain.Quantity(50), trades[1].Quantity)
	assert.Equal(t, domain.Decimal(982), trades[1].Price)

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), taker.Status().String())
	assert.Equal(t, domain.Quantity(0), taker.Remaining())

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker1.Status().String())
	assert.Equal(t, domain.AllocationStatusPartial.String(), maker2.AllocationStatus().String())
	assert.Equal(t, domain.OrderStatusActive.String(), maker2.Status().String())
	assert.Equal(t, domain.Quantity(150), maker2.Remaining())
	repo.AssertExpectations(t)
}

func Test_MarketTakerPartiallyFilledIsKilled(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	maker, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "limit", 100, 981, 100, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	orderBook := []*domain.Order{maker}

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return(orderBook, nil).Maybe()

	// Market order: no price, only quantity matters.
	taker, err := domain.NewOrder(uuid.New(), uuid.New(), "bid", "market", 150, 0, 0, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)
	assert.Nil(t, taker.Price())

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Len(t, trades, 1)

	assert.Equal(t, domain.Quantity(100), trades[0].Quantity)
	assert.Equal(t, domain.Decimal(981), trades[0].Price)

	assert.Equal(t, domain.Quantity(100), taker.AllocationQty())
	assert.Equal(t, domain.Quantity(50), taker.Remaining())
	// Leftover quantity on a market order: killed, not left resting.
	assert.Equal(t, domain.OrderStatusKilled.String(), taker.Status().String())

	assert.Equal(t, domain.OrderStatusFullyAllocated.String(), maker.Status().String())
	repo.AssertExpectations(t)
}

func Test_MarketTakerEmptyBookIsKilled(t *testing.T) {
	repo := mocks.NewMockOrderRepo()
	internalCrossing := domain.AllowCrossingPolicy{}
	me := New(repo, internalCrossing)

	repo.On("GetOrdersToTrade", mock.Anything, mock.AnythingOfType("*domain.Order")).
		Return([]*domain.Order{}, nil).Maybe()

	taker, err := domain.NewOrder(uuid.New(), uuid.New(), "ask", "market", 100, 0, 0, domain.Decimal(100), domain.Decimal(1), expiration)
	assert.NoError(t, err)

	matches, trades, err := me.ExecuteIncomingOrder(t.Context(), taker)

	assert.NoError(t, err)
	assert.Empty(t, matches)
	assert.Empty(t, trades)
	assert.Equal(t, domain.Quantity(100), taker.Remaining())
	assert.Equal(t, domain.OrderStatusKilled.String(), taker.Status().String())
	repo.AssertExpectations(t)
}
