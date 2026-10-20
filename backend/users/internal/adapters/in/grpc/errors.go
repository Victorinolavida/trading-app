package grpc

import (
	"errors"

	"github.com/victorinolavida/auth/internal/domainerrors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, domainerrors.ErrInvalidEmail),
		errors.Is(err, domainerrors.ErrNameTooShort),
		errors.Is(err, domainerrors.ErrNameTooLong),
		errors.Is(err, domainerrors.ErrPasswordTooShort),
		errors.Is(err, domainerrors.ErrPasswordTooLong):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domainerrors.ErrEmailAlreadyRegister):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domainerrors.ErrInvalidCredentials):
		// keep generic to avoid user enumeration
		return status.Error(codes.Unauthenticated, "invalid credentials")

	case errors.Is(err, domainerrors.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")

	default:
		// don't leak internal errors to the client
		return status.Error(codes.Internal, "internal error")
	}
}
