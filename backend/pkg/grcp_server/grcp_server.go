package grcpserver

import (
	"context"
	"log/slog"
	"net"
	"strconv"

	"github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"google.golang.org/grpc"
)

var (
	defaultProtocol = "tcp"
	defaultPort     = 8100
	defaultHost     = "localhost"
)

type GRCPServerInterface interface {
	Listen() net.Listener
}

type GRCPServer struct {
	conf config.GRPCServer
	svr  *grpc.Server
}

var _ GRCPServerInterface = (*GRCPServer)(nil)

func NewGRCPServer(conf config.GRPCServer, interceptors ...grpc.UnaryServerInterceptor) *GRCPServer {
	base := []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			// you can keep your existing base logic here
			return handler(ctx, req)
		},
	}
	all := append(base, interceptors...)
	svr := grpc.NewServer(grpc.ChainUnaryInterceptor(all...))
	return &GRCPServer{
		conf: conf,
		svr:  svr,
	}
}

func (s *GRCPServer) Server() *grpc.Server {
	return s.svr
}

func (s *GRCPServer) Listen() net.Listener {
	protocol := s.conf.Protocol
	if protocol == "" {
		protocol = defaultProtocol
	}

	grcpServerAddress := formatGRPCAddres(s.conf.Host, s.conf.Port)
	l, err := net.Listen(protocol, grcpServerAddress)
	if err != nil {
		logger.Fatal("error listening for grcp server", slog.Any("error", err))
	}
	slog.Info("GRCP server started", slog.String("addr", grcpServerAddress))
	return l
}

func formatGRPCAddres(host string, port int) string {
	if host == "" {
		host = defaultHost
	}
	if port == 0 {
		port = defaultPort
	}
	return host + ":" + strconv.Itoa(port)
}
