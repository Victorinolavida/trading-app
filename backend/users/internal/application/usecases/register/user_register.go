package register

import (
	"context"

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
	user, err := domain.NewUser(
		command.Email,
		command.FirstName,
		command.LastName,
		command.Password,
	)
	if err != nil {
		return Result{}, err
	}
	exist, err := u.repo.ExistByEmail(ctx, user.Email())
	if err != nil {
		return Result{}, err
	}
	if exist {
		return Result{}, domainerrors.ErrEmailAlreadyRegister
	}

	err = u.repo.Save(ctx, user)
	if err != nil {
		return Result{}, err
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
