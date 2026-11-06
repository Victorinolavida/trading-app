package domain

import (
	"errors"
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/victorinolavida/securities/internal/domainerrors"
)

type (
	SecurityType string
	Currency     string
)

const (
	MBond   SecurityType = "M"
	UdiBond SecurityType = "S"

	MXN Currency = "MXN"
	UDI Currency = "UDI"
)

var ErrInvalidNumericNumber = errors.New("invalid numeric value")

func (s SecurityType) String() string { return string(s) }

func (s SecurityType) validate() error {
	if !slices.Contains([]SecurityType{MBond, UdiBond}, s) {
		return domainerrors.ErrInvalidSecurityType
	}
	return nil
}

func (c Currency) validate() error {
	if !slices.Contains([]Currency{MXN, UDI}, c) {
		return domainerrors.ErrInvalidCurrency
	}
	return nil
}
func (c Currency) String() string { return string(c) }

func NewSecurityType(val string) (SecurityType, error) {
	securityType := SecurityType(val)
	if err := securityType.validate(); err != nil {
		return SecurityType(""), err
	}
	return securityType, nil
}

func NewCurrency(val string) (Currency, error) {
	currency := Currency(val)
	if err := currency.validate(); err != nil {
		return Currency(""), err
	}
	return currency, nil
}

func NumericToFloat(n pgtype.Numeric) (float64, error) {
	val, err := n.Float64Value()
	if err != nil {
		return float64(0), err
	}
	if !val.Valid {
		return float64(0), ErrInvalidNumericNumber
	}
	return val.Float64, nil
}
