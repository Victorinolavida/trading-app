// Package session checks both token signatures and the current user session version.
package session

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/jwt"
)

type UserLookup interface {
	Whoami(context.Context, dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error)
}

type Verifier struct {
	tokens ports.TokenVerifier
	users  UserLookup
}

func NewVerifier(tokens ports.TokenVerifier, users UserLookup) *Verifier {
	return &Verifier{tokens: tokens, users: users}
}

func (v *Verifier) Verify(ctx context.Context, raw string) (ports.Claims, error) {
	claims, err := v.tokens.Verify(ctx, raw)
	if err != nil {
		return ports.Claims{}, err
	}
	user, err := v.users.Whoami(ctx, dtos.WhoamiUserRequest{UserID: claims.UserID})
	if err != nil {
		return ports.Claims{}, err
	}
	if user.SessionVersion != claims.SessionVersion {
		return ports.Claims{}, jwt.ErrUnauthenticated
	}
	return claims, nil
}
