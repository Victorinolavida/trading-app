package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/victorinolavida/pkg/config"
)

func testConf() *config.CookieConfig {
	return &config.CookieConfig{
		Name:     "auth_token",
		CSRFName: "csrf_token",
		Path:     "/",
		Secure:   true,
		SameSite: "None",
		TTL:      3600,
		Secret:   "12345678901234567890123456789012", // 32 bytes
	}
}

func TestIssueAndVerify(t *testing.T) {
	conf := testConf()
	token, err := Issue(conf)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v0/test", nil)
	req.AddCookie(&http.Cookie{Name: conf.CSRFName, Value: token})
	req.Header.Set("X-CSRF-Token", token)

	if err := Verify(req, conf); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerifyRejectsMissingCookie(t *testing.T) {
	conf := testConf()
	req := httptest.NewRequest(http.MethodPost, "/api/v0/test", nil)
	if err := Verify(req, conf); err != ErrNoCSRFToken {
		t.Fatalf("expected ErrNoCSRFToken, got %v", err)
	}
}

func TestVerifyRejectsMissingHeader(t *testing.T) {
	conf := testConf()
	token, _ := Issue(conf)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/test", nil)
	req.AddCookie(&http.Cookie{Name: conf.CSRFName, Value: token})
	if err := Verify(req, conf); err != ErrNoCSRFToken {
		t.Fatalf("expected ErrNoCSRFToken, got %v", err)
	}
}

func TestVerifyRejectsBadSignature(t *testing.T) {
	conf := testConf()
	req := httptest.NewRequest(http.MethodPost, "/api/v0/test", nil)
	req.AddCookie(&http.Cookie{Name: conf.CSRFName, Value: "abc.def"})
	req.Header.Set("X-CSRF-Token", "abc.def")
	if err := Verify(req, conf); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifyRejectsHeaderMismatch(t *testing.T) {
	conf := testConf()
	token, _ := Issue(conf)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/test", nil)
	req.AddCookie(&http.Cookie{Name: conf.CSRFName, Value: token})
	req.Header.Set("X-CSRF-Token", token+"x")
	if err := Verify(req, conf); err == nil {
		t.Fatal("expected error")
	}
}
