package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"users/config"
	"users/internal/adapters/secondary_adapter/jwt"
	"users/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testSecret = "test-secret-0123456789-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJ"

type userRepoMock struct{ mock.Mock }

// var _ ports.UserRepository = (*userRepoMock)(nil)

func (r *userRepoMock) Save(ctx context.Context, u domain.User) error {
	args := r.Called(ctx, u)
	return args.Error(0)
}

func (r *userRepoMock) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	args := r.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (r *userRepoMock) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	args := r.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (r *userRepoMock) ExistByEmail(ctx context.Context, email domain.Email) (bool, error) {
	args := r.Called(ctx, email)
	return args.Get(0).(bool), args.Error(1)
}

func (r *userRepoMock) Update(ctx context.Context, user domain.User) error {
	args := r.Called(ctx, user)
	return args.Error(0)
}

func TestRegisterUser_DbFails(t *testing.T) {
	repo := new(userRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := NewUserUsecase(repo, jwtIssuer)

	cmd := RegisterCommand{
		Email:     "john@example.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.User")).Return(errors.New("db error"))

	result, err := handler.Register(t.Context(), cmd)

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUser_InvalidEmail(t *testing.T) {
	repo := new(userRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := NewUserUsecase(repo, jwtIssuer)

	cmd := RegisterCommand{
		Email:     "john@.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	result, err := handler.Register(t.Context(), cmd)

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUser_HappyPath(t *testing.T) {
	repo := new(userRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := NewUserUsecase(repo, jwtIssuer)
	cmd := RegisterCommand{
		Email:     "john@example.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.User")).Return(nil)

	result, err := handler.Register(t.Context(), cmd)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.User.ID)
	assert.Equal(t, cmd.Email, result.User.Email)
	repo.AssertExpectations(t)
}
