package whoami

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type Usecase struct {
	repo ports.UserRepository
}

func New(repo ports.UserRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	userUUID, err := uuid.Parse(cmd.UserID)
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
