// Command app runs the whole backend as a modular monolith: the HTTP gateway
// and the users, securities and orders bounded contexts live in one process and
// talk to each other through their public api packages.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

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
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	log := logger.NewLogger("app")

	cfg, err := conf.NewConfig[appconfig.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}
	if err := cfg.Validate(); err != nil {
		logger.Fatal("invalid config", slog.Any("error", err))
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stopSignalLog := context.AfterFunc(rootCtx, func() {
		log.Info("shutdown signal received")
	})
	defer stopSignalLog()

	usersPool, err := database.NewPool(rootCtx, cfg.Database.For(cfg.Database.Users))
	if err != nil {
		logger.Fatal("error connecting to users database", slog.Any("error", err))
	}
	defer usersPool.Close()

	securitiesPool, err := database.NewPool(rootCtx, cfg.Database.For(cfg.Database.Securities))
	if err != nil {
		logger.Fatal("error connecting to securities database", slog.Any("error", err))
	}
	defer securitiesPool.Close()

	ordersPool, err := database.NewPool(rootCtx, cfg.Database.For(cfg.Database.Orders))
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

	log.Info("starting modules")

	g, gctx := errgroup.WithContext(rootCtx)
	g.Go(func() error { return usersModule.Start(gctx) })
	g.Go(func() error { return securitiesModule.Start(gctx) })
	g.Go(func() error { return ordersModule.Start(gctx) })
	g.Go(func() error { return server.Run(gctx) })

	runErr := g.Wait()
	if runErr != nil {
		log.Error("module stopped with error", slog.Any("error", runErr))
	}

	log.Info("modules stopped")

	closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Close(); err != nil {
		log.Error("error closing gateway", slog.Any("error", err))
	}
	if err := ordersModule.Close(closeCtx); err != nil {
		log.Error("error closing orders module", slog.Any("error", err))
	}
	if err := securitiesModule.Close(closeCtx); err != nil {
		log.Error("error closing securities module", slog.Any("error", err))
	}
	if err := usersModule.Close(closeCtx); err != nil {
		log.Error("error closing users module", slog.Any("error", err))
	}

	log.Info("resources closed")

	usersPool.Close()
	securitiesPool.Close()
	ordersPool.Close()

	log.Info("exit")
	return runErr
}
