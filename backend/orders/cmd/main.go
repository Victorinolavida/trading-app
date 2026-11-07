package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/victorinolavida/orders/config"
	grpcAdapter "github.com/victorinolavida/orders/internal/adapters/in/grpc"
	"github.com/victorinolavida/orders/internal/adapters/out/grpc/securities"
	"github.com/victorinolavida/orders/internal/adapters/out/grpc/users"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres"
	matchingengine "github.com/victorinolavida/orders/internal/application/services/matching-engine"
	"github.com/victorinolavida/orders/internal/application/usecases/active_orders"
	"github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	"github.com/victorinolavida/orders/internal/application/usecases/create_order"
	"github.com/victorinolavida/orders/internal/application/usecases/list_orders"
	"github.com/victorinolavida/orders/internal/domain"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grpcserver "github.com/victorinolavida/pkg/grpc_server"
	"github.com/victorinolavida/pkg/logger"
	orderpb "github.com/victorinolavida/proto/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	logs := logger.NewLogger()

	cfg, err := conf.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading configs", slog.Any("error", err))
	}

	pool, err := database.NewPool(context.Background(), cfg.Database)
	if err != nil {
		logger.Fatal("error connecting to db", slog.Any("error", err))
	}

	orderRepo := postgres.NewOrderRepository(pool)
	tradeRepo := postgres.NewTradeRepository(pool)
	txManager := postgres.NewTxManager(pool)
	usersConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", cfg.Services.Users.Host, cfg.Services.Users.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("error connecting to users service", slog.Any("error", err))
	}
	securitiesConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", cfg.Services.Securities.Host, cfg.Services.Securities.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("error connecting to securities service", slog.Any("error", err))
	}

	userSvc := users.New(usersConn)
	securitySvc := securities.New(securitiesConn)

	var internalCrossinPolicy domain.InternalCrossingPolicy = domain.SameTraderCrossingPolicy{Action: domain.CrossingCancelTaker}
	if cfg.AllowInternalCrossing {
		logs.Warn("internal crossing is DISABLE, be sure that this is expected in configs for production env")
		internalCrossinPolicy = domain.AllowCrossingPolicy{}
	}

	matchingEngine := matchingengine.New(orderRepo, internalCrossinPolicy)
	createUC := create_order.New(orderRepo, tradeRepo, securitySvc, userSvc, matchingEngine, txManager)
	cancelUC := cancel_order.New(orderRepo)
	activeUc := active_orders.New(orderRepo)
	listUc := list_orders.New(orderRepo)

	grpcServer := grpcserver.NewGRPCServer(cfg.GRPC)
	handler := grpcAdapter.NewOrderGRPCHandler(createUC, cancelUC, activeUc, listUc)
	orderpb.RegisterOrderServiceServer(grpcServer.Server(), handler)

	reflection.Register(grpcServer.Server())

	listener := grpcServer.Listen()

	go func() {
		if err := grpcServer.Server().Serve(listener); err != nil {
			logger.Fatal("grpc server error", slog.Any("error", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	grpcServer.Server().GracefulStop()
	logs.Info("orders server shutdown successfully")
}
