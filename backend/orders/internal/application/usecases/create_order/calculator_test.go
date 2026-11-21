package create_order

import (
	"context"
	"github.com/google/uuid"
	"testing"
)

type calculatorStub struct {
	field, value, result string
	calls                int
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
