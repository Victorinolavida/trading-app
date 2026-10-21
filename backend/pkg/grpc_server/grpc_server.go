package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"strconv"

	"github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	defaultProtocol = "tcp"
	defaultPort     = 3000
	defaultHost     = "localhost"
)

type GRPCServerInterface interface {
	Listen() net.Listener
}

type GRPCServer struct {
	conf config.GRPCServer
	svr  *grpc.Server
}

var _ GRPCServerInterface = (*GRPCServer)(nil)

func NewGRPCServer(conf config.GRPCServer, interceptors ...grpc.UnaryServerInterceptor) *GRPCServer {
	// Create health service
	healthService := health.NewServer()

	base := []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			// you can keep your existing base logic here
			return handler(ctx, req)
		},
	}

	all := append(base, interceptors...)
	all = append(all, recoveryInterceptor())

	// Register the health service with the gRPC server
	server := grpc.NewServer(grpc.ChainUnaryInterceptor(all...))
	grpc_health_v1.RegisterHealthServer(server, healthService)
	// Set the service health status
	healthService.SetServingStatus(conf.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	return &GRPCServer{
		conf: conf,
		svr:  server,
	}
}

func (s *GRPCServer) Server() *grpc.Server {
	return s.svr
}

func (s *GRPCServer) Listen() net.Listener {
	protocol := s.conf.Protocol
	if protocol == "" {
		protocol = defaultProtocol
	}

	grpcServerAddress := formatGRPCAddress(s.conf.Host, s.conf.Port)
	l, err := net.Listen(protocol, grpcServerAddress)
	if err != nil {
		logger.Fatal("error listening for grpc server", slog.Any("error", err))
	}
	slog.Info("GRPC server started", slog.String("addr", grpcServerAddress))
	return l
}

func formatGRPCAddress(host string, port int) string {
	if host == "" {
		host = defaultHost
	}
	if port == 0 {
		port = defaultPort
	}
	return host + ":" + strconv.Itoa(port)
}
