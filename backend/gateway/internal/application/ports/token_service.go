package ports

import (
	"context"

	"github.com/victorinolavida/pkg/jwt"
)

// Claims is the identity carried by the token.
type Claims = jwt.Claims

type TokenVerifier interface {
	Verify(ctx context.Context, raw string) (Claims, error)
}
