package session

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/jwt"
)

type tokenStub struct {
	claims ports.Claims
	err    error
}

func (s tokenStub) Verify(context.Context, string) (ports.Claims, error) { return s.claims, s.err }

type userStub struct {
	user dtos.WhoamiUserResponse
	err  error
}

func (s userStub) Whoami(context.Context, dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error) {
	return s.user, s.err
}

func TestVerifier(t *testing.T) {
	claims := ports.Claims{UserID: "user", SessionVersion: 2}
	failure := errors.New("service unavailable")
	for _, tc := range []struct {
		name    string
		tokens  tokenStub
		users   userStub
		wantErr error
	}{
		{"valid", tokenStub{claims: claims}, userStub{user: dtos.WhoamiUserResponse{SessionVersion: 2}}, nil},
		{"revoked", tokenStub{claims: claims}, userStub{user: dtos.WhoamiUserResponse{SessionVersion: 3}}, jwt.ErrUnauthenticated},
		{"invalid token", tokenStub{err: jwt.ErrUnauthenticated}, userStub{}, jwt.ErrUnauthenticated},
		{"lookup failure", tokenStub{claims: claims}, userStub{err: failure}, failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewVerifier(tc.tokens, tc.users).Verify(context.Background(), "token")
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, claims, got)
		})
	}
}
