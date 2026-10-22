package whoami

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/victorinolavida/auth/internal/application/usecases/mocks"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

func Test_WhoamiEmptyID(t *testing.T) {
	repo := new(mocks.UserRepoMock)

	handler := New(repo)

	result, err := handler.Execute(t.Context(), Command{
		UserID: "",
	})

	assert.Error(t, err)
	assert.Empty(t, result.UserID)
	assert.Empty(t, result.FirstName)
	repo.AssertExpectations(t)
}

func Test_WhoamiBadID(t *testing.T) {
	repo := new(mocks.UserRepoMock)
	handler := New(repo)

	result, err := handler.Execute(t.Context(), Command{
		UserID: "someid",
	})

	assert.Error(t, err)
	assert.Empty(t, result.UserID)
	assert.Empty(t, result.FirstName)
	repo.AssertExpectations(t)
}

func Test_WhoamiUserNotFound(t *testing.T) {
	user, err := domain.Reconstruct(
		uuid.MustParse("08f4583e-ec0c-4ac4-b773-8aafcdb1014c"),
		"jhon.rod@gmail.com",
		"jhon",
		"Rod",
		[]byte("hardcoredpassword"),
		time.Now(),
		time.Now(),
	)

	assert.NoError(t, err)

	repo := new(mocks.UserRepoMock)
	handler := New(repo)

	repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(domain.User{}, domainerrors.ErrUserNotFound)

	result, err := handler.Execute(t.Context(), Command{
		UserID: user.ID().String(),
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerrors.ErrUnauthenticated)
	assert.Empty(t, result.Email)
	assert.Empty(t, result.UserID)
	assert.Empty(t, result.FirstName)
	assert.Empty(t, result.LastName)
	repo.AssertExpectations(t)
}

func Test_WhoamiHappyPath(t *testing.T) {
	user, err := domain.Reconstruct(
		uuid.MustParse("08f4583e-ec0c-4ac4-b773-8aafcdb1014c"),
		"jhon.rod@gmail.com",
		"jhon",
		"Rod",
		[]byte("hardcoredpassword"),
		time.Now(),
		time.Now(),
	)

	assert.NoError(t, err)

	repo := new(mocks.UserRepoMock)
	handler := New(repo)

	repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(user, nil)

	result, err := handler.Execute(t.Context(), Command{
		UserID: user.ID().String(),
	})

	assert.NoError(t, err)
	assert.Equal(t, user.Email().String(), result.Email)
	assert.Equal(t, user.ID().String(), result.UserID)
	assert.Equal(t, user.FirstName().String(), result.FirstName)
	assert.Equal(t, user.LastName().String(), result.LastName)
	repo.AssertExpectations(t)
}
