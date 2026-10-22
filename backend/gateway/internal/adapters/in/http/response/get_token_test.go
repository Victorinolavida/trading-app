package response

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTokenFromRequestBearerTakesPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer token-from-header")
	req.AddCookie(&http.Cookie{Name: "session", Value: "token-from-cookie"})

	token, source, err := GetTokenFromRequest(req, "session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-from-header" {
		t.Fatalf("expected header token, got %q", token)
	}
	if source != SourceBearer {
		t.Fatalf("expected source bearer, got %q", source)
	}
}

func TestGetTokenFromRequestCookieFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "token-from-cookie"})

	token, source, err := GetTokenFromRequest(req, "session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-from-cookie" {
		t.Fatalf("expected cookie token, got %q", token)
	}
	if source != SourceCookie {
		t.Fatalf("expected source cookie, got %q", source)
	}
}

func TestGetTokenFromRequestMalformedHeaderDoesNotFallBack(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc123")
	req.AddCookie(&http.Cookie{Name: "session", Value: "token-from-cookie"})

	token, source, err := GetTokenFromRequest(req, "session")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrNoCredential) {
		t.Fatalf("expected ErrNoCredential, got %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
	if source != SourceInvalid {
		t.Fatalf("expected source invalid, got %q", source)
	}
}

func TestGetTokenFromRequestNoCredentials(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, _, err := GetTokenFromRequest(req, "session")
	if !errors.Is(err, ErrNoCredential) {
		t.Fatalf("expected ErrNoCredential, got %v", err)
	}
}
