package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/victorinolavida/orders/config"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres"
	orderproducer "github.com/victorinolavida/orders/internal/adapters/out/producer/rabbitmq"
	redisadapter "github.com/victorinolavida/orders/internal/adapters/out/redis"
	orderexpiration "github.com/victorinolavida/orders/internal/application/workers/order_expiration"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
	cfgpkg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	"github.com/victorinolavida/pkg/logger"
	"github.com/victorinolavida/pkg/redis"
)

func main() {
	logs := logger.NewLogger()

	cfg, err := cfgpkg.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading configs", slog.Any("error", err))
	}
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := database.NewPool(context.Background(), cfg.Database)
	if err != nil {
		logger.Fatal("error connecting to db", slog.Any("error", err))
	}
	defer pool.Close()

	exchange := cfg.Publisher.Exchange
	if exchange == "" {
		logger.Fatal("publisher exchange is required")
	}

	redisClient, err := redis.NewClient(context.Background(), cfg.Redis)
	if err != nil {
		logger.Fatal("error with redis", slog.Any("error", err))
	}
	defer redisClient.Close()

	memoryPort, err := redisadapter.New(context.Background(), cfg.Redis)
	if err != nil {
		logger.Fatal("error with redis memory adapter", slog.Any("error", err))
	}

	rabbitClient, err := rabbitmq.NewClientFromConfig(cfg.RabbitMQ)
	if err != nil {
		logger.Fatal("error with rabbitmq adapter", slog.Any("error", err))
	}
	defer rabbitClient.Close()

	orderRepo := postgres.NewOrderRepository(pool)
	txManager := postgres.NewTxManager(pool)
	eventPublisher, err := orderproducer.NewProducer(memoryPort, rabbitClient, exchange)
	if err != nil {
		logger.Fatal("error creating order event producer", slog.Any("error", err))
	}

	worker := orderexpiration.NewOrderWorker(orderRepo, txManager, eventPublisher, time.Second*10)
	if err := worker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logs.Error("worker stopped with error", slog.Any("error", err))
		return // lets defers run
	}
	<-ctx.Done()
}
