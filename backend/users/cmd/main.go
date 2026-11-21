package main

import (
	"context"
	"log/slog"
	"os/signal"
	"strings"
	"syscall"

	"github.com/victorinolavida/auth/config"
	grpcAdapter "github.com/victorinolavida/auth/internal/adapters/in/grpc"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/metadata"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/requestid"
	emailmock "github.com/victorinolavida/auth/internal/adapters/out/email/mock"
	"github.com/victorinolavida/auth/internal/adapters/out/postgres"
	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/passwordreset"
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
	whoamiHandler := whoami.New(userRepo)
	var authHandler *grpcAdapter.AuthGRPCHandler
	if cfg.PasswordReset.Enabled {
		if cfg.PasswordReset.Provider != "mock" || !strings.EqualFold(cfg.Env, "development") {
			logger.Fatal("unsupported password reset email provider or environment")
		}
		resetHandler, err := passwordreset.New(userRepo, postgres.NewPasswordResetRepository(pool), &emailmock.Sender{}, cfg.PasswordReset.URL, cfg.PasswordReset.TTL)
		if err != nil {
			logger.Fatal("invalid password reset configuration", slog.Any("error", err))
		}
		authHandler = grpcAdapter.NewAuthGRPCHandler(registerHandler, loginHandler, whoamiHandler, resetHandler)
	} else {
		authHandler = grpcAdapter.NewAuthGRPCHandler(registerHandler, loginHandler, whoamiHandler, nil)
	}

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
