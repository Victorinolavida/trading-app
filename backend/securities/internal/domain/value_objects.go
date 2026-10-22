package domain

import (
	"slices"

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
