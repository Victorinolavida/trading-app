package domainerrors

import (
	"errors"

	"github.com/victorinolavida/pkg/jwt"
)

var (
	ErrInvalidEmail         = errors.New("invalid email")
	ErrNameTooShort         = errors.New("is too short")
	ErrNameTooLong          = errors.New("is too long")
	ErrPasswordTooShort     = errors.New("password too short")
	ErrPasswordTooLong      = errors.New("password too long")
	ErrUserNotFound         = errors.New("user not found")
	ErrUnauthenticated      = jwt.ErrUnauthenticated
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailAlreadyRegister = errors.New("email already register")
)
