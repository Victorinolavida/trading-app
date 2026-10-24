package domainerrors

import "errors"

var (
	ErrInvalidMaturityDate      = errors.New("maturity date must be greater than issue date")
	ErrInvalidISIN              = errors.New("ISIN must containt 12 characteres")
	ErrNominalValueLessThanZero = errors.New("nominal value less than 0")
	ErrFxLessThanZero           = errors.New("fx is less than 0")
	ErrEmptySerie               = errors.New("a serie value is required")
	ErrEmptyName                = errors.New("name is required")
	ErrInvalidSecurityType      = errors.New("invalid security type")
	ErrInvalidCurrency          = errors.New("invalid currency")
	ErrSecurityNotFound         = errors.New("security not found")
)
