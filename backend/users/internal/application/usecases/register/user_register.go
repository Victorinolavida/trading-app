package usecases

import (
	"context"

	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type UserRegister struct {
	repo   ports.UserRepository
	issuer ports.TokenIssuer
}

func NewUserUsecase(repo ports.UserRepository, issuer ports.TokenIssuer) *UserRegister {
	return &UserRegister{repo: repo, issuer: issuer}
}

type RegisterCommand struct {
	Email     string
	FirstName string
	LastName  string
	Password  string
}

type RegisterResult struct {
	User struct {
		ID        string
		Email     string
		FirstName string
		LastName  string
	}
	Token string
}

func (u *UserRegister) Register(ctx context.Context, command RegisterCommand) (RegisterResult, error) {
	user, err := domain.NewUser(
		command.Email,
		command.FirstName,
		command.LastName,
		command.Password,
	)
	if err != nil {
		return RegisterResult{}, err
	}
	exist, err := u.repo.ExistByEmail(ctx, user.Email())
	if err != nil {
		return RegisterResult{}, err
	}
	if exist {
		return RegisterResult{}, domainerrors.ErrEmailAlreadyRegister
	}

	err = u.repo.Save(ctx, user)
	if err != nil {
		return RegisterResult{}, err
	}
	token, err := u.issuer.Issue(ctx, ports.Claims{
		UserID:    user.ID().String(),
		FirstName: user.FirstName().String(),
		LastName:  user.LastName().String(),
	})
	if err != nil {
		return RegisterResult{}, err
	}
	return RegisterResult{
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
