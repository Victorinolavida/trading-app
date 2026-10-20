package login

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorinolavida/auth/config"
	"github.com/victorinolavida/auth/internal/adapters/out/jwt"
	"github.com/victorinolavida/auth/internal/application/usecases/mocks"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

const testSecret = "test-secret-0123456789-abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJ"

func TestLoginUser_BadEmail(t *testing.T) {
	repo := new(mocks.UserRepoMock)
	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	handler := NewUserUsecase(repo, jwtIssuer)

	result, err := handler.Handle(context.Background(), LoginCommand{
		Password: "password123",
		Email:    "email@.com",
	})

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	assert.Empty(t, result.User.ID)
}

func TestLoginUser_EmailDoesNotExist(t *testing.T) {
	repo := new(mocks.UserRepoMock)
	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})

	assert.NoError(t, err, "error creating new token sevice")
	repo.On("GetByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(domain.User{}, domainerrors.ErrUserNotFound)
	handler := NewUserUsecase(repo, jwtIssuer)

	result, err := handler.Handle(context.Background(), LoginCommand{
		Password: "password123",
		Email:    "test@gmail.com",
	})

	assert.ErrorIs(t, err, domainerrors.ErrInvalidCredentials)
	assert.Empty(t, result.Token)
	assert.Empty(t, result.User.ID)

	repo.AssertExpectations(t)
}

func TestLoginUser_PasswordNotMatch(t *testing.T) {
	repo := new(mocks.UserRepoMock)
	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})
	assert.NoError(t, err, "error creating new token sevice")

	plainPass := "super_secure_password"
	pass, err := domain.NewPassword(plainPass)
	assert.NoError(t, err, "invalid create a test password")

	mockUser, err := domain.Reconstruct(
		uuid.New(),
		"test@gmail.com",
		"Jhon",
		"Rod",
		pass.Hash(),
		time.Now(),
		time.Now(),
	)
	assert.NoError(t, err, "invalid reconstructur a user")

	repo.On("GetByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(mockUser, nil)
	handler := NewUserUsecase(repo, jwtIssuer)

	result, err := handler.Handle(context.Background(), LoginCommand{
		Password: plainPass + "123", // diferent to password from above
		Email:    "test@gmail.com",
	})

	assert.Error(t, err)
	assert.Empty(t, result.Token)
	assert.Empty(t, result.User.ID)

	repo.AssertExpectations(t)
}

func TestLoginUser_HappyPath(t *testing.T) {
	repo := new(mocks.UserRepoMock)
	jwtIssuer, err := jwt.NewService(config.TokenConfig{
		Secret:   testSecret,
		Issuer:   "test",
		Audience: "test-issuer",
		TTL:      15 * time.Minute,
	})
	assert.NoError(t, err, "error creating new token sevice")

	plainPass := "super_secure_password"
	pass, err := domain.NewPassword(plainPass)
	assert.NoError(t, err, "failed to create a test password")

	userID := uuid.New()
	mockUser, err := domain.Reconstruct(
		userID,
		"test@gmail.com",
		"John",
		"Doe",
		pass.Hash(),
		time.Now(),
		time.Now(),
	)
	assert.NoError(t, err, "failed to reconstruct a user")

	repo.On("GetByEmail", mock.Anything, mock.AnythingOfType("domain.Email")).Return(mockUser, nil)
	handler := NewUserUsecase(repo, jwtIssuer)

	result, err := handler.Handle(context.Background(), LoginCommand{
		Password: plainPass,
		Email:    "test@gmail.com",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, userID.String(), result.User.ID)
	assert.Equal(t, "test@gmail.com", result.User.Email)
	assert.Equal(t, "John", result.User.FirstName)
	assert.Equal(t, "Doe", result.User.LastName)

	repo.AssertExpectations(t)
}
