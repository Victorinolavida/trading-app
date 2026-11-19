package domain

import (
	"errors"
	"fmt"
	"strings"
)

type (
	Quantity int64
	Decimal  int64
)

const (
	PriceScale      int64    = 1_000 // three decimals
	DecimalScale    int64    = 1_000_000
	MaxPrice        int64    = 1_000 * PriceScale
	MaxQty          Quantity = 1_000_000_000
	MaxDecimalValue int64    = int64(MaxPrice) * int64(MaxQty)
)

func (d Decimal) Value() int64 {
	return int64(d)
}

func (q Quantity) Value() int64 {
	return int64(q)
}

var (
	ErrInvalidNumericNumber    = errors.New("invalid numeric value")
	ErrInvalidAllocationStatus = errors.New("invalid allocation status")
)

func NewQuantity(qty int) (Quantity, error) {
	if qty <= 0 {
		return Quantity(0), ErrInvalidNumericNumber
	}
	q := Quantity(qty)
	if q > MaxQty {
		return Quantity(0), ErrInvalidNumericNumber
	}
	return q, nil
}

func NewDecimal(val int64) (Decimal, error) {
	if val < 0 || val > MaxDecimalValue {
		return Decimal(0), ErrInvalidNumericNumber
	}
	return Decimal(val), nil
}

// FormatDecimal returns the string representation of d, or "" when d is nil.
func FormatDecimal(d *Decimal) string {
	if d == nil {
		return ""
	}
	return d.String()
}

func (d Decimal) String() string {
	v := int64(d)
	neg := v < 0
	if neg {
		v = -v
	}
	whole, frac := v/int64(DecimalScale), v%int64(DecimalScale)
	s := fmt.Sprintf("%d", whole)
	if frac != 0 {
		s += "." + strings.TrimRight(fmt.Sprintf("%06d", frac), "0")
	}
	if neg {
		return "-" + s
	}
	return s
}
