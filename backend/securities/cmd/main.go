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

	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"github.com/victorinolavida/securities/config"
)

const shutdownTimeout = 15 * time.Second

func main() {
	log := logger.NewLogger()

	conf, err := cfg.NewConfig[config.Config]()
	if err != nil {
		logger.Fatal("error loading config", slog.Any("error", err))
	}

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("error shutting down http server", slog.Any("error", err))
	}

	log.Info("server shutdown successfully")
}
