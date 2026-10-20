package login

import (
	"context"
	"errors"

	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type UserLoginHandler struct {
	repo   ports.UserRepository
	issuer ports.TokenIssuer
}

func NewUserUsecase(repo ports.UserRepository, issuer ports.TokenIssuer) *UserLoginHandler {
	return &UserLoginHandler{repo: repo, issuer: issuer}
}

func (u *UserLoginHandler) Handle(ctx context.Context, command LoginCommand) (LoginResult, error) {
	email, err := domain.NewEmail(command.Email)
	if err != nil {
		return LoginResult{}, err
	}
	user, err := u.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return LoginResult{}, domainerrors.ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	passwordMatch := domain.VerifyPassword(command.Password, string(user.Hash()))
	if !passwordMatch {
		return LoginResult{}, domainerrors.ErrInvalidCredentials
	}

	token, err := u.issuer.Issue(ctx, ports.Claims{
		UserID:    user.ID().String(),
		LastName:  user.LastName().String(),
		FirstName: user.FirstName().String(),
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
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
