package login

import (
	"context"
	"errors"

	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type Usecase struct {
	repo   ports.UserRepository
	issuer ports.TokenIssuer
}

func New(repo ports.UserRepository, issuer ports.TokenIssuer) *Usecase {
	return &Usecase{repo: repo, issuer: issuer}
}

func (u *Usecase) Execute(ctx context.Context, command Command) (Result, error) {
	email, err := domain.NewEmail(command.Email)
	if err != nil {
		return Result{}, err
	}
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return Result{}, domainerrors.ErrInvalidCredentials
		}
		return Result{}, err
	}

	passwordMatch := domain.VerifyPassword(command.Password, string(user.Hash()))
	if !passwordMatch {
		return Result{}, domainerrors.ErrInvalidCredentials
	}

	token, err := u.issuer.Issue(ctx, ports.Claims{
		UserID: user.ID().String(),
	})
	if err != nil {
		return Result{}, err
	}
	return Result{
		User: struct {
			ID        string
			Email     string
			FirstName string
			LastName  string
		}{
			ID:        user.ID().String(),
			FirstName: user.FirstName().String(),
			LastName:  user.LastName().String(),
			Email:     user.Email().String(),
		},
		Token: token,
	}, nil
}
