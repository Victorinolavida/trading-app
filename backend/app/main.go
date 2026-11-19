// Command app runs the whole backend as a modular monolith: the HTTP gateway
// and the users, securities and orders bounded contexts live in one process and
// talk to each other through their public api packages.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	appconfig "github.com/victorinolavida/app/config"
	usersapi "github.com/victorinolavida/auth/api"
	gatewayapi "github.com/victorinolavida/gateway/api"
	ordersapi "github.com/victorinolavida/orders/api"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	"github.com/victorinolavida/pkg/logger"
	securitiesapi "github.com/victorinolavida/securities/api"
)

const shutdownTimeout = 30 * time.Second

func main() {
	log := logger.NewLogger("app")

	cfg, err := conf.NewConfig[appconfig.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatal("invalid config", slog.Any("error", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	usersPool, err := database.NewPool(ctx, cfg.Database.For(cfg.Database.Users))
	if err != nil {
		logger.Fatal("error connecting to users database", slog.Any("error", err))
	}
	defer usersPool.Close()

	securitiesPool, err := database.NewPool(ctx, cfg.Database.For(cfg.Database.Securities))
	if err != nil {
		logger.Fatal("error connecting to securities database", slog.Any("error", err))
	}
	defer securitiesPool.Close()

	ordersPool, err := database.NewPool(ctx, cfg.Database.For(cfg.Database.Orders))
	if err != nil {
		logger.Fatal("error connecting to orders database", slog.Any("error", err))
	}
	defer ordersPool.Close()

	usersModule, err := usersapi.New(usersPool, cfg.Token)
	if err != nil {
		logger.Fatal("error wiring users module", slog.Any("error", err))
	}

	securitiesModule := securitiesapi.New(securitiesPool)

	ordersModule, err := ordersapi.New(ordersapi.Deps{
		Pool:                  ordersPool,
		Redis:                 cfg.Redis,
		RabbitMQ:              cfg.RabbitMQ,
		Exchange:              cfg.Publisher.Exchange,
		AllowInternalCrossing: cfg.AllowInternalCrossing,
		Users:                 userProvider{module: usersModule},
		Securities:            securityProvider{module: securitiesModule},
	})
	if err != nil {
		logger.Fatal("error wiring orders module", slog.Any("error", err))
	}

	gatewayCfg := cfg.Gateway()
	server, err := gatewayapi.New(&gatewayCfg, usersModule, securitiesModule, ordersModule)
	if err != nil {
		logger.Fatal("error wiring gateway", slog.Any("error", err))
	}

	workerCtx, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()

	go func() {
		if err := ordersModule.StartWorker(workerCtx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("order expiration worker stopped", slog.Any("error", err))
		}
	}()

	go func() {
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("error shutting down http server", slog.Any("error", err))
	}

	stopWorker()

	if err := ordersModule.Close(); err != nil {
		slog.Error("error closing orders module", slog.Any("error", err))
	}
	if err := server.Close(); err != nil {
		slog.Error("error closing gateway", slog.Any("error", err))
	}

	log.Info("monolith shutdown successfully")
}
