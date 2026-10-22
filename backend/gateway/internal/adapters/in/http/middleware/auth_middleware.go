package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/ports"
)

type authCtxKey string

const authorizationKey authCtxKey = "auth_ctx_key"

func AuthMiddleware(tokenService ports.TokenVerifier) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := response.GetTokenFromRequest(r)
			if err != nil {
				slog.Error("error extracting token from request", slog.Any("error", err))
				response.RenderUnauthorized(w)
				return
			}

			claims, err := tokenService.Verify(r.Context(), token)
			if err != nil {
				slog.Error("error verifying token", slog.Any("error", err))
				response.RenderUnauthorized(w)
				return
			}

			ctx := context.WithValue(r.Context(), authorizationKey, claims)
			r = r.WithContext(ctx)
			h.ServeHTTP(w, r)
		})
	}
}

func GetClaims(ctx context.Context) (ports.Claims, error) {
	claims, ok := ctx.Value(authorizationKey).(ports.Claims)
	if !ok {
		return ports.Claims{}, errors.New("unauthorized")
	}
	return claims, nil
}
