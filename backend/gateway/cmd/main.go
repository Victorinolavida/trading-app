package main

import (
	"context"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"github.com/victorinolavida/gateway/config"
	authhandler "github.com/victorinolavida/gateway/internal/adapters/in/http/auth_handler"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	ordershandler "github.com/victorinolavida/gateway/internal/adapters/in/http/orders_handler"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	securityhandler "github.com/victorinolavida/gateway/internal/adapters/in/http/security_handler"
	httpserver "github.com/victorinolavida/gateway/internal/adapters/in/http/server"
	authservice "github.com/victorinolavida/gateway/internal/adapters/out/grpc/auth_service"
	ordersservice "github.com/victorinolavida/gateway/internal/adapters/out/grpc/orders_service"
	securityservice "github.com/victorinolavida/gateway/internal/adapters/out/grpc/security_service"
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

	if err := conf.Validate(); err != nil {
		logger.Fatal("invalid config", slog.Any("error", err))
	}

	if conf.Cookie.TTL > conf.Token.TTL {
		slog.Warn("cookie ttl outlives token ttl, the cookie will outlive the token it carries",
			slog.Duration("cookie_ttl", conf.Cookie.TTL),
			slog.Duration("token_ttl", conf.Token.TTL),
		)
	}

	httpServer := httpserver.NewHttpServer(*conf)

	authService, err := authservice.NewAuthService(conf.AuthService)
	if err != nil {
		logger.Fatal("error in auth service", slog.Any("error", err))
	}

	secService, err := securityservice.NewSecurityService(conf.SecuritiesService)
	if err != nil {
		logger.Fatal("error in securities service", slog.Any("error", err))
	}

	ordService, err := ordersservice.NewOrdersService(conf.OrdersService)
	if err != nil {
		logger.Fatal("error in orders service", slog.Any("error", err))
	}

	authHandler, err := authhandler.NewAuthHandler(conf, authService, response.NewValidator())
	if err != nil {
		logger.Fatal("error creating auth handler", slog.Any("error", err))
	}

	secHandler, err := securityhandler.NewSecurityHandler(conf, secService, response.NewValidator())
	if err != nil {
		logger.Fatal("error creating security handler", slog.Any("error", err))
	}

	ordHandler, err := ordershandler.NewOrdersHandler(conf, ordService, response.NewValidator())
	if err != nil {
		logger.Fatal("error creating orders handler", slog.Any("error", err))
	}

	modules := []module.Module{
		authHandler,
		secHandler,
		ordHandler,
	}

	for _, mod := range modules {
		httpServer.MountModule(apiPrefix, mod)
	}

	// Walk and log all registered routes
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("Route registered",
			slog.String("method", method),
			slog.String("route", route),
			slog.Int("middlewares_count", len(middlewares)),
		)
		return nil
	}

	if err := chi.Walk(httpServer.ChiRoutes(), walkFunc); err != nil {
		slog.Error("Failed to walk routes", slog.Any("error", err))
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
	if err := secService.Close(); err != nil {
		slog.Error("error closing securities service connection", slog.Any("error", err))
	}
	if err := ordService.Close(); err != nil {
		slog.Error("error closing orders service connection", slog.Any("error", err))
	}

	log.Info("gateway server shutdown successfully")
}
