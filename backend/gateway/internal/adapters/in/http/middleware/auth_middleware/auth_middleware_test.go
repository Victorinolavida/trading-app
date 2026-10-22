package authmiddleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/csrf"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/config"
)

type stubVerifier struct {
	err error
}

func (s *stubVerifier) Verify(ctx context.Context, raw string) (ports.Claims, error) {
	if s.err != nil {
		return ports.Claims{}, s.err
	}
	return ports.Claims{UserID: "user-123"}, nil
}

func testCookieConf() *config.CookieConfig {
	return &config.CookieConfig{
		Name:     "auth_token",
		CSRFName: "csrf_token",
		Path:     "/",
		Secret:   "12345678901234567890123456789012",
		Secure:   true,
		SameSite: "None",
		TTL:      time.Hour,
	}
}

func TestAuthMiddlewareBearerSkipsCSRF(t *testing.T) {
	r := chi.NewRouter()
	r.With(AuthMiddleware(&stubVerifier{}, testCookieConf())).Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer anytoken")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareCookieRequiresCSRF(t *testing.T) {
	r := chi.NewRouter()
	r.With(AuthMiddleware(&stubVerifier{}, testCookieConf())).Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "token"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestAuthMiddlewareCookieAcceptsWithCSRF(t *testing.T) {
	conf := testCookieConf()
	csrfToken, err := csrf.Issue(conf)
	if err != nil {
		t.Fatal(err)
	}

	r := chi.NewRouter()
	r.With(AuthMiddleware(&stubVerifier{}, conf)).Post("/test", func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaims(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if claims.UserID != "user-123" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: conf.Name, Value: "token"})
	req.AddCookie(&http.Cookie{Name: conf.CSRFName, Value: csrfToken})
	req.Header.Set("X-CSRF-Token", csrfToken)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestAuthMiddlewareCookieGETSkipsCSRF(t *testing.T) {
	r := chi.NewRouter()
	r.With(AuthMiddleware(&stubVerifier{}, testCookieConf())).Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "auth_token", Value: "token"})
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
