package passwordreset

import (
	"context"
	"crypto/sha256"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	emailmock "github.com/victorinolavida/auth/internal/adapters/out/email/mock"
	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/application/usecases/mocks"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type resetRepo struct {
	digest []byte
	hash   []byte
}

func (r *resetRepo) Store(_ context.Context, _ uuid.UUID, digest []byte, _ time.Time) error {
	r.digest = append([]byte(nil), digest...)
	return nil
}

func (r *resetRepo) Redeem(_ context.Context, digest, hash []byte) error {
	if string(digest) != string(r.digest) || r.digest == nil {
		return ports.ErrInvalidResetToken
	}
	r.digest = nil
	r.hash = hash
	return nil
}

func TestRequestAndConfirm(t *testing.T) {
	ctx := context.Background()
	users := &mocks.UserRepoMock{}
	user, err := domain.Reconstruct(uuid.New(), "user@example.com", "First", "Last", nil, time.Now(), time.Now())
	require.NoError(t, err)
	address, err := domain.NewEmail("user@example.com")
	require.NoError(t, err)
	users.On("GetByEmail", ctx, address).Return(user, nil)
	resets := &resetRepo{}
	inbox := &emailmock.Sender{}
	usecase, err := New(users, resets, inbox, "https://example.com/reset-password", time.Hour)
	require.NoError(t, err)
	require.NoError(t, usecase.Request(ctx, address.String()))
	messages := inbox.Messages()
	require.Len(t, messages, 1)
	link, err := url.Parse(strings.TrimPrefix(messages[0].Text, "Use this link to reset your password: "))
	require.NoError(t, err)
	token := link.Query().Get("token")
	digest := sha256.Sum256([]byte(token))
	require.Equal(t, digest[:], resets.digest)
	require.NoError(t, usecase.Confirm(ctx, token, "new-password"))
	require.True(t, domain.VerifyPassword("new-password", string(resets.hash)))
	require.ErrorIs(t, usecase.Confirm(ctx, token, "new-password"), ports.ErrInvalidResetToken)
}

func TestUnknownEmailDoesNotSend(t *testing.T) {
	ctx := context.Background()
	users := &mocks.UserRepoMock{}
	address, err := domain.NewEmail("unknown@example.com")
	require.NoError(t, err)
	users.On("GetByEmail", ctx, address).Return(domain.User{}, domainerrors.ErrUserNotFound)
	inbox := &emailmock.Sender{}
	usecase, err := New(users, &resetRepo{}, inbox, "https://example.com/reset-password", time.Hour)
	require.NoError(t, err)
	require.NoError(t, usecase.Request(ctx, address.String()))
	require.Empty(t, inbox.Messages())
}
