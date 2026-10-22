package whoami

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type Usecase struct {
	repo     ports.UserRepository
	verifier ports.TokenVerifier
}

func New(repo ports.UserRepository, verifier ports.TokenVerifier) *Usecase {
	return &Usecase{repo: repo, verifier: verifier}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	claims, err := u.verifier.Verify(ctx, cmd.Token)
	if err != nil {
		return Result{}, domainerrors.ErrUnauthenticated
	}

	userUUID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return Result{}, domainerrors.ErrUnauthenticated
	}
	user, err := u.repo.GetByID(ctx, userUUID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return Result{}, domainerrors.ErrUnauthenticated
		}
		return Result{}, err
	}

	return Result{
		Email:     user.Email().String(),
		FirstName: user.FirstName().String(),
		LastName:  user.LastName().String(),
		UserID:    user.ID().String(),
	}, nil
}
