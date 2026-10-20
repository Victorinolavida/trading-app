package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
)

type UserRepoMock struct{ mock.Mock }

var _ ports.UserRepository = (*UserRepoMock)(nil)

func (r *UserRepoMock) Save(ctx context.Context, u domain.User) error {
	args := r.Called(ctx, u)
	return args.Error(0)
}

func (r *UserRepoMock) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := r.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (r *UserRepoMock) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	args := r.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (r *UserRepoMock) ExistByEmail(ctx context.Context, email domain.Email) (bool, error) {
	args := r.Called(ctx, email)
	return args.Get(0).(bool), args.Error(1)
}

func (r *UserRepoMock) Update(ctx context.Context, user domain.User) error {
	args := r.Called(ctx, user)
	return args.Error(0)
}
