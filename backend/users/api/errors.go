package api

import "github.com/victorinolavida/auth/internal/domainerrors"

// Validation and authentication errors re-exported for callers that translate
// module failures at their own boundary.
var (
	ErrInvalidEmail         = domainerrors.ErrInvalidEmail
	ErrNameTooShort         = domainerrors.ErrNameTooShort
	ErrNameTooLong          = domainerrors.ErrNameTooLong
	ErrPasswordTooShort     = domainerrors.ErrPasswordTooShort
	ErrPasswordTooLong      = domainerrors.ErrPasswordTooLong
	ErrUserNotFound         = domainerrors.ErrUserNotFound
	ErrUnauthenticated      = domainerrors.ErrUnauthenticated
	ErrInvalidCredentials   = domainerrors.ErrInvalidCredentials
	ErrEmailAlreadyRegister = domainerrors.ErrEmailAlreadyRegister
)
