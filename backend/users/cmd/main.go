package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/victorinolavida/auth/config"
	grpcAdapter "github.com/victorinolavida/auth/internal/adapters/in/grpc"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/metadata"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/requestid"
	"github.com/victorinolavida/auth/internal/adapters/out/postgres"
	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/register"
	"github.com/victorinolavida/auth/internal/application/usecases/whoami"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grpcserver "github.com/victorinolavida/pkg/grpc_server"
	"github.com/victorinolavida/pkg/jwt"
	"github.com/victorinolavida/pkg/logger"
	authpb "github.com/victorinolavida/proto/auth"
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

	// user repo
	userRepo := postgres.NewUserRepository(pool)
	tokenServ, err := jwt.NewService(cfg.TokenConfig)
	if err != nil {
		logger.Fatal("error creating new token service", slog.Any("error", err))
	}

	grpcServer := grpcserver.NewGRPCServer(cfg.GRPC, requestid.UnaryServerInterceptor, metadata.UnaryServerInterceptor, grpcAdapter.ErrorMappingInterceptor)

	registerHandler := register.New(userRepo, tokenServ)
	loginHandler := login.New(userRepo, tokenServ)
	whoamiHandler := whoami.New(userRepo, tokenServ)
	authHandler := grpcAdapter.NewAuthGRPCHandler(registerHandler, loginHandler, whoamiHandler)

	authpb.RegisterAuthServiceServer(grpcServer.Server(), authHandler)

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
	logs.Info("auth server shutdown successfully")
}
