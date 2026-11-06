package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/victorinolavida/securities/config"
	grpcAdapter "github.com/victorinolavida/securities/internal/adapters/in/grcp"
	"github.com/victorinolavida/securities/internal/adapters/out/postgres"
	getenablesecurities "github.com/victorinolavida/securities/internal/application/usecases/get_enable_securities"
	getsecurities "github.com/victorinolavida/securities/internal/application/usecases/get_securities"
	getsecuritybyparams "github.com/victorinolavida/securities/internal/application/usecases/get_security_by_params"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grpcserver "github.com/victorinolavida/pkg/grpc_server"
	"github.com/victorinolavida/pkg/logger"
	securitypb "github.com/victorinolavida/proto/security"
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

	repo := postgres.NewSecurityRepo(pool)

	getEnable := getenablesecurities.New(repo)
	getByParams := getsecurities.New(repo)
	getByParam := getsecuritybyparams.New(repo)

	grpcServer := grpcserver.NewGRPCServer(cfg.GRPC)
	handler := grpcAdapter.NewSecurityGRPCHandler(getEnable, getByParams, getByParam)
	securitypb.RegisterAuthServiceServer(grpcServer.Server(), handler)

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
	logs.Info("securities server shutdown successfully")
}

