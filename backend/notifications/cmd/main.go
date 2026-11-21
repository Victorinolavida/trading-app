package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/victorinolavida/pkg/broker/rabbitmq"
	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grpcserver "github.com/victorinolavida/pkg/grpc_server"
	"github.com/victorinolavida/pkg/logger"
	pb "github.com/victorinolavida/proto/notification"

	"module/config"
	grpcadapter "module/internal/adapters/in/grpc"
	"module/internal/adapters/out/postgres"
	"module/internal/consumer"
)

const shutdownTimeout = 15 * time.Second

func main() {
	log := logger.NewLogger()

	conf, err := cfg.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	if conf.Consumer.Exchange == "" || conf.Consumer.Queue == "" {
		logger.Fatal("consumer exchange and queue are required")
	}
	pool, err := database.NewPool(ctx, conf.Database)
	if err != nil {
		logger.Fatal("error connecting to database", slog.Any("error", err))
	}
	defer pool.Close()
	broker, err := rabbitmq.NewClientFromConfig(conf.RabbitMQ)
	if err != nil {
		logger.Fatal("error connecting to rabbitmq", slog.Any("error", err))
	}
	defer broker.Close()
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		consumer.New(broker, postgres.New(pool), conf.Consumer.Exchange, conf.Consumer.Queue).Run(ctx)
	}()
	grpcServer := grpcserver.NewGRPCServer(conf.GRPC)
	pb.RegisterNotificationServiceServer(grpcServer.Server(), grpcadapter.New(postgres.New(pool)))
	listener := grpcServer.Listen()
	go func() {
		if err := grpcServer.Server().Serve(listener); err != nil {
			log.Error("grpc server stopped", slog.Any("error", err))
			stop()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              net.JoinHostPort(conf.Server.Host, strconv.Itoa(conf.Server.Port)),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("http server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()
	grpcServer.Server().GracefulStop()
	<-consumerDone

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("error shutting down http server", slog.Any("error", err))
	}

	log.Info("server shutdown successfully")
}
