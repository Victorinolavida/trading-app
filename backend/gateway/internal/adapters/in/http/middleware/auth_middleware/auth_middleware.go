package authmiddleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/csrf"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/config"
)

type authCtxKey string

const authorizationKey authCtxKey = "auth_ctx_key"

func AuthMiddleware(tokenService ports.TokenVerifier, cookieConf *config.CookieConfig) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, source, err := response.GetTokenFromRequest(r, cookieConf.Name)
			if err != nil {
				if !errors.Is(err, response.ErrNoCredential) {
					slog.Warn("error extracting token from request", slog.Any("error", err))
				} else {
					slog.Debug("request without credentials")
				}
				response.RenderUnauthorized(w)
				return
			}

			// Only the cookie path is forgeable by a third-party site, so only the
			// cookie path needs the CSRF check.
			if source == response.SourceCookie && isUnsafeMethod(r.Method) {
				if err := csrf.Verify(r, cookieConf); err != nil {
					slog.Warn("csrf verification failed",
						slog.Any("error", err),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)
					response.RenderForbidden(w)
					return
				}
			}

			claims, err := tokenService.Verify(r.Context(), token)
			if err != nil {
				slog.Warn("error verifying token", slog.Any("error", err))
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

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}
