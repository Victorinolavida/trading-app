package domain_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/orders/internal/domain"
)

func TestDecimal_String(t *testing.T) {
	tests := []struct {
		name string
		val  domain.Decimal
		want string
	}{
		{"zero", 0, "0"},
		{"one micro unit", 1, "0.000001"},
		{"one thousandth", 1_000, "0.001"},
		{"one unit", 1_000_000, "1"},
		{"trims trailing zeros", 1_500_000, "1.5"},
		{"price 98.12", 98_120_000, "98.12"},
		{"keeps inner zeros", 98_120_100, "98.1201"},
		{"whole hundred", 100_000_000, "100"},
		{"max fractional digits", 999_999, "0.999999"},
		{"large value", 123_456_789_012, "123456.789012"},
		{"negative", -98_120_000, "-98.12"},
		{"negative micro unit", -1, "-0.000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.val.String())
		})
	}
}
