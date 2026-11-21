package main

import (
	"context"
	"github.com/victorinolavida/calculator/api"
	"github.com/victorinolavida/calculator/config"
	"github.com/victorinolavida/calculator/internal/security"
	"github.com/victorinolavida/calculator/internal/transport"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	grpcserver "github.com/victorinolavida/pkg/grpc_server"
	"github.com/victorinolavida/pkg/logger"
	calculatorpb "github.com/victorinolavida/proto/calculator"
	securitypb "github.com/victorinolavida/proto/security"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := conf.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading configs", slog.Any("error", err))
	}
	client, err := grpcclient.NewGRPCClient(cfg.SecuritiesService)
	if err != nil {
		logger.Fatal("error creating securities client", slog.Any("error", err))
	}
	defer client.Close()
	pool, err := database.NewPool(context.Background(), cfg.Database)
	if err != nil {
		logger.Fatal("error connecting to securities database", slog.Any("error", err))
	}
	defer pool.Close()
	module, err := api.New(api.Deps{Security: security.Provider{Client: securitypb.NewAuthServiceClient(client.Client)}, Udi: security.UDIProvider{Pool: pool}})
	if err != nil {
		logger.Fatal("error wiring calculator", slog.Any("error", err))
	}
	server := grpcserver.NewGRPCServer(cfg.GRPC)
	calculatorpb.RegisterCalculatorServiceServer(server.Server(), &transport.Server{Module: module})
	reflection.Register(server.Server())
	listener := server.Listen()
	go func() {
		if err := server.Server().Serve(listener); err != nil {
			logger.Fatal("calculator grpc server error", slog.Any("error", err))
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()
	server.Server().GracefulStop()
}
