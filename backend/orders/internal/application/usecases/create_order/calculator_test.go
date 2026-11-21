package create_order

import (
	"context"
	"github.com/google/uuid"
	"testing"
	"time"

	"github.com/victorinolavida/orders/internal/domain"
)

type calculatorStub struct {
	field, value, result string
	calls                int
}

func TestCalculatedZeroYieldIsRetainedInOrder(t *testing.T) {
	calc := &calculatorStub{result: "0.0000004"}
	u := Usecase{calculator: calc}
	cmd := Command{TraderID: uuid.New(), SecurityID: uuid.New(), Direction: "bid", OrderType: "limit", Quantity: 1, Price: 100000000, ExpiresAt: time.Hour}
	if err := u.completeLimitOrder(context.Background(), &cmd); err != nil {
		t.Fatal(err)
	}
	order, err := domain.NewOrder(cmd.TraderID, cmd.SecurityID, cmd.Direction, cmd.OrderType, int(cmd.Quantity), cmd.Price, cmd.Yield, 1000000, 1000000, cmd.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	if calc.calls != 1 || order.Yield() == nil || *order.Yield() != 0 {
		t.Fatalf("calculated zero yield was not retained: calls=%d yield=%v", calc.calls, order.Yield())
	}
}

func (c *calculatorStub) Calculate(_ context.Context, _, _, field, value string) (string, error) {
	c.field, c.value = field, value
	c.calls++
	return c.result, nil
}
func TestCompleteLimitOrder(t *testing.T) {
	for _, tc := range []struct {
		name                               string
		price, yield, wantPrice, wantYield int64
		field, value, result               string
		calls                              int
	}{
		{"price", 98125000, 0, 98125000, 81235, "clean_price", "98.125", "0.0812346", 1},
		{"yield", 0, 81235, 98125000, 81235, "yield", "0.081235", "98.125", 1},
		{"both", 98125000, 81235, 98125000, 81235, "", "", "", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calc := &calculatorStub{result: tc.result}
			u := Usecase{calculator: calc}
			cmd := Command{SecurityID: uuid.New(), OrderType: "limit", Price: tc.price, Yield: tc.yield}
			if err := u.completeLimitOrder(context.Background(), &cmd); err != nil {
				t.Fatal(err)
			}
			if cmd.Price != tc.wantPrice || cmd.Yield != tc.wantYield || calc.calls != tc.calls || calc.field != tc.field || calc.value != tc.value {
				t.Fatalf("got command %+v, calculator %+v", cmd, calc)
			}
		})
	}
}
