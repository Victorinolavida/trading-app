package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(args ...string) *slog.Logger {
	var register slog.Handler
	if len(args) == 2 && isProd(args[1]) {
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

	log := slog.New(register)

	slog.SetDefault(log)
	return log
}

func isProd(env string) bool {
	return strings.ToLower(env) == "production"
}

func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	panic(1)
}
