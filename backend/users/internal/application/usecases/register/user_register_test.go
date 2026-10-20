package register

import (
	"errors"
	"testing"
	"time"

	"github.com/victorinolavida/auth/config"
	"github.com/victorinolavida/auth/internal/adapters/out/jwt"
	"github.com/victorinolavida/auth/internal/application/usecases/mocks"
	"github.com/victorinolavida/auth/internal/domainerrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const testSecret = "test-secret-0123456789-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJ"

func TestRegisterUser_DbFails(t *testing.T) {
	repo := new(mocks.UserRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := New(repo, jwtIssuer)

	cmd := Command{
		Email:     "john@example.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	repo.On("ExistByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(false, errors.New("db error"))

	result, err := handler.Execute(t.Context(), cmd)

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUser_InvalidEmail(t *testing.T) {
	repo := new(mocks.UserRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := New(repo, jwtIssuer)

	cmd := Command{
		Email:     "john@.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	result, err := handler.Execute(t.Context(), cmd)

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUser_DuplicatedEmail(t *testing.T) {
	repo := new(mocks.UserRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := New(repo, jwtIssuer)
	cmd := Command{
		Email:     "john@example.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	repo.On("ExistByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(true, nil)

	result, err := handler.Execute(t.Context(), cmd)

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrEmailAlreadyRegister)
	assert.Empty(t, result.Token)
	repo.AssertExpectations(t)
}

func TestRegisterUser_HappyPath(t *testing.T) {
	repo := new(mocks.UserRepoMock)

	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := New(repo, jwtIssuer)
	cmd := Command{
		Email:     "john@example.com",
		Password:  "secret123",
		FirstName: "John",
		LastName:  "Doe",
	}

	repo.On("Save", mock.Anything, mock.AnythingOfType("domain.User")).Return(nil)
	repo.On("ExistByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(false, nil)

	result, err := handler.Execute(t.Context(), cmd)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.User.ID)
	assert.Equal(t, cmd.Email, result.User.Email)
	repo.AssertExpectations(t)
}
