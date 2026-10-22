package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"github.com/victorinolavida/gateway/config"
	authhandler "github.com/victorinolavida/gateway/internal/adapters/in/http/auth_handler"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	httpserver "github.com/victorinolavida/gateway/internal/adapters/in/http/server"
	authservice "github.com/victorinolavida/gateway/internal/adapters/out/grpc/auth_service"
)

const (
	defaultTimeout = 15 * time.Second
	apiPrefix      = "/api/v0"
)

func main() {
	log := logger.NewLogger()

	conf, err := cfg.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}

	httpServer := httpserver.NewHttpServer(*conf)

	authService, err := authservice.NewAuthService(conf.AuthService)
	if err != nil {
		logger.Fatal("error in auth service", slog.Any("error", err))
	}

	authHandler, err := authhandler.NewAuthHandler(conf, authService, response.NewValidator())
	if err != nil {
		logger.Fatal("error creating auth handler", slog.Any("error", err))
	}

	modules := []module.Module{
		authHandler,
	}

	for _, mod := range modules {
		httpServer.MountModule(apiPrefix, mod)
	}

	go func() {
		if err := httpServer.Start(); err != nil && http.ErrServerClosed != err {
			logger.Fatal("closing http server", slog.Any("error", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("http error shutdown", slog.Any("error", err))
	}

	if err := authService.Close(); err != nil {
		slog.Error("error closing auth service connection", slog.Any("error", err))
	}

	log.Info("gateway server shutdown successfully")
}
