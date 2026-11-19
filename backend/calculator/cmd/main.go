package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/victorinolavida/calculator/api"
	"github.com/victorinolavida/calculator/config"
	pkgconfig "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"
)

func main() {
	if _, err := pkgconfig.NewConfig[config.Config](); err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}

	module := api.New(nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := module.Start(ctx); err != nil {
		logger.Fatal("calculator module error", slog.Any("error", err))
	}
	if err := module.Close(ctx); err != nil {
		slog.Error("calculator module close error", slog.Any("error", err))
	}
}
