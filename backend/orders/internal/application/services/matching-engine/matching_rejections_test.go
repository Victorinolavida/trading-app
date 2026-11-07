package matchingengine

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/victorinolavida/orders/internal/domain"
)

func Test_canExecuteFunction(t *testing.T) {
	t.Run("orders with the same direction", func(t *testing.T) {
		taker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"bid",
			"limit",
			10,
			10,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")
		maker := &taker // clone order

		rejection := canExecute(taker, *maker)
		assert.Equal(t, RejectSameDirection, rejection)
	})

	t.Run("limit orders with no prices coss bid", func(t *testing.T) {
		taker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"bid",
			"limit",
			10,
			10,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		maker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"ask",
			"limit",
			10,
			11,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		rejection := canExecute(taker, maker)
		assert.Equal(t, RejectPriceDoesNotCross, rejection)
	})

	t.Run("limit orders with no prices coss ask", func(t *testing.T) {
		taker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"ask",
			"limit",
			10,
			11,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		maker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"bid",
			"limit",
			10,
			10,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		rejection := canExecute(taker, maker)
		assert.Equal(t, RejectPriceDoesNotCross, rejection)
	})

	t.Run("market orders with prices coss ask", func(t *testing.T) {
		taker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"ask",
			"market",
			1_000,
			0,
			0,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		maker, err := domain.NewOrder(
			uuid.New(),
			uuid.New(),
			"bid",
			"limit",
			1_000,
			10,
			10,
			domain.Decimal(100),
			domain.Decimal(1),
		)
		assert.NoError(t, err, "no error build order")

		rejection := canExecute(taker, maker)
		assert.Equal(t, MatchOK, rejection)
	})
}
