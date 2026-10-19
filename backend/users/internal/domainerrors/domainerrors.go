package domainerrors

import "errors"

var (
	ErrInvalidEmail = errors.New("invalid email")
	ErrNameTooShort = errors.New("is too short")
	ErrNameTooLong  = errors.New("is too short")
)
