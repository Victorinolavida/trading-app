package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/victorinolavida/auth/config"
	grpcAdapter "github.com/victorinolavida/auth/internal/adapters/in/grpc"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/metadata"
	"github.com/victorinolavida/auth/internal/adapters/in/grpc/requestid"
	emailmock "github.com/victorinolavida/auth/internal/adapters/out/email/mock"
	emailsmtp "github.com/victorinolavida/auth/internal/adapters/out/email/smtp"
	"github.com/victorinolavida/auth/internal/adapters/out/postgres"
	"github.com/victorinolavida/auth/internal/application/ports"
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
		var sender ports.EmailSender
		switch cfg.PasswordReset.Provider {
		case "mock":
			if !strings.EqualFold(cfg.Env, "development") {
				logger.Fatal("mock email provider is development-only")
			}
			sender = emailmock.NewSender(func(message ports.Email) {
				logs.Info("development email captured", slog.String("to", message.To), slog.String("subject", message.Subject), slog.String("body", message.Text))
			})
		case "mailpit":
			if !strings.EqualFold(cfg.Env, "development") {
				logger.Fatal("mailpit email provider is development-only")
			}
			sender, err = emailsmtp.NewSender(cfg.PasswordReset.Mailpit.Host, cfg.PasswordReset.Mailpit.Port, cfg.PasswordReset.Mailpit.From)
			if err != nil {
				logger.Fatal("invalid mailpit configuration", slog.Any("error", err))
			}
		case "smtp":
			if !strings.HasPrefix(strings.ToLower(cfg.PasswordReset.URL), "https://") {
				logger.Fatal("SMTP password reset requires an HTTPS reset URL")
			}
			c := cfg.PasswordReset.SMTP
			if v := os.Getenv("RESET_SMTP_USERNAME"); v != "" {
				c.Username = v
			}
			if v := os.Getenv("RESET_SMTP_PASSWORD"); v != "" {
				c.Password = v
			}
			sender, err = emailsmtp.NewAuthenticatedSender(c.Host, c.Port, c.From, c.Username, c.Password)
			if err != nil {
				logger.Fatal("invalid SMTP configuration", slog.Any("error", err))
			}
		default:
			logger.Fatal("unsupported password reset email provider")
		}
		resetHandler, err := passwordreset.New(userRepo, postgres.NewPasswordResetRepository(pool), sender, cfg.PasswordReset.URL, cfg.PasswordReset.TTL)
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
