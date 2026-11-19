package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// NewLogger builds the default logger. service is used as the fallback when the
// originating module cannot be derived from the log call site. env optionally
// selects the environment ("production" enables JSON output).
func NewLogger(service string, env ...string) *slog.Logger {
	var register slog.Handler
	if len(env) > 0 && isProd(env[0]) {
		register = slog.NewJSONHandler(os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			})
	} else {
		register = slog.NewTextHandler(os.Stdout,
			&slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			})
	}

	log := slog.New(newServiceHandler(register, service))

	slog.SetDefault(log)
	return log
}

func isProd(env string) bool {
	return strings.ToLower(env) == "production"
}

// Fatal logs the message and terminates the process with exit code 1.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

// serviceHandler wraps a slog.Handler and attributes every record to the
// module that emitted it, falling back to a fixed service name.
type serviceHandler struct {
	inner   slog.Handler
	service string
}

func newServiceHandler(inner slog.Handler, service string) *serviceHandler {
	return &serviceHandler{inner: inner, service: service}
}

func (h *serviceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *serviceHandler) Handle(ctx context.Context, record slog.Record) error {
	name := serviceFromPC(record.PC)
	if name == "" {
		name = h.service
	}

	if name == "" {
		return h.inner.Handle(ctx, record)
	}

	cloned := record.Clone()
	cloned.AddAttrs(slog.String("service", name))
	return h.inner.Handle(ctx, cloned)
}

func (h *serviceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &serviceHandler{inner: h.inner.WithAttrs(attrs), service: h.service}
}

func (h *serviceHandler) WithGroup(name string) slog.Handler {
	return &serviceHandler{inner: h.inner.WithGroup(name), service: h.service}
}

func serviceFromPC(pc uintptr) string {
	if pc == 0 {
		return ""
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	return serviceFromName(fn.Name())
}

func serviceFromName(name string) string {
	switch {
	case strings.Contains(name, "/gateway/"):
		return "gateway"
	case strings.Contains(name, "/auth/"), strings.Contains(name, "/users/"):
		return "users"
	case strings.Contains(name, "/securities/"):
		return "securities"
	case strings.Contains(name, "/orders/"):
		return "orders"
	case strings.Contains(name, "/app/"):
		return "app"
	default:
		return ""
	}
}
