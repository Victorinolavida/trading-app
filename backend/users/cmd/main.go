package main

import (
	"context"
	"github.com/victorinolavida/auth/config"
	"github.com/victorinolavida/auth/internal/adapters/secondary_adapter/postgres"
	conf "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/database"
	grcpserver "github.com/victorinolavida/pkg/grcp_server"
	"github.com/victorinolavida/pkg/logger"
	"log/slog"
)

func main() {
	logger.NewLogger()

	conf, err := conf.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading configs", slog.Any("error", err))
	}

	pool, err := database.NewPool(context.Background(), conf.Database)
	if err != nil {
		logger.Fatal("error connecting to db", slog.Any("error", err))
	}

	// user repo
	_ = postgres.NewUserRepository(pool)

	grcpServer := grcpserver.NewGRCPServer(conf.Grcp)
	grcpServer.Listen()
}
