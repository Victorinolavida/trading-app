package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func newTestLogger(fallback string, buf *bytes.Buffer) *slog.Logger {
	inner := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(newServiceHandler(inner, fallback))
}

func TestServiceHandlerFallback(t *testing.T) {
	var buf bytes.Buffer
	newTestLogger("fallback", &buf).Info("hello")

	out := buf.String()
	if !strings.Contains(out, "service=fallback") {
		t.Fatalf("expected fallback service attribute, got %q", out)
	}
}

func TestServiceHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	newTestLogger("fallback", &buf).With("component", "test").Info("hello")

	out := buf.String()
	if !strings.Contains(out, "service=fallback") {
		t.Fatalf("expected service attribute, got %q", out)
	}
	if !strings.Contains(out, "component=test") {
		t.Fatalf("expected preserved attribute, got %q", out)
	}
}

func TestServiceHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	newTestLogger("fallback", &buf).WithGroup("group").Info("hello", "key", "value")

	out := buf.String()
	if !strings.Contains(out, "group.service=fallback") {
		t.Fatalf("expected grouped service attribute, got %q", out)
	}
	if !strings.Contains(out, "group.key=value") {
		t.Fatalf("expected grouped attribute, got %q", out)
	}
}

func TestServiceFromName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"github.com/victorinolavida/gateway/internal/http.Handler", "gateway"},
		{"github.com/victorinolavida/auth/internal/user.Service", "users"},
		{"github.com/victorinolavida/users/internal/user.Service", "users"},
		{"github.com/victorinolavida/securities/internal/security.Service", "securities"},
		{"github.com/victorinolavida/orders/internal/order.Service", "orders"},
		{"github.com/victorinolavida/app/internal/wiring.Build", "app"},
		{"main.main", ""},
	}

	for _, tt := range tests {
		if got := serviceFromName(tt.name); got != tt.want {
			t.Errorf("serviceFromName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
