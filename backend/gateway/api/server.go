// Package api is the public composition boundary of the gateway.
//
// It mounts the HTTP feature modules and wires them to the in-process bounded
// contexts. A single process can therefore serve the whole backend without a
// gRPC hop between modules.
package api

import (
	"context"

	usersapi "github.com/victorinolavida/auth/api"
	"github.com/victorinolavida/gateway/config"
	authhandler "github.com/victorinolavida/gateway/internal/adapters/in/http/auth_handler"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	ordershandler "github.com/victorinolavida/gateway/internal/adapters/in/http/orders_handler"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	securityhandler "github.com/victorinolavida/gateway/internal/adapters/in/http/security_handler"
	httpserver "github.com/victorinolavida/gateway/internal/adapters/in/http/server"
	ordersapi "github.com/victorinolavida/orders/api"
	brokerrabbit "github.com/victorinolavida/pkg/broker/rabbitmq"
	securitiesapi "github.com/victorinolavida/securities/api"
)

const apiPrefix = "/api/v0"

// Server is the HTTP entrypoint of the monolith.
type Server struct {
	http   httpserver.ServerInterface
	rabbit *brokerrabbit.Client
}

// New wires the HTTP server and its feature modules on top of the in-process
// bounded contexts.
func New(
	cfg *config.Config,
	users *usersapi.Module,
	securities *securitiesapi.Module,
	orders *ordersapi.Module,
) (*Server, error) {
	httpServer := httpserver.NewHttpServer(*cfg)

	authHandler, err := authhandler.NewAuthHandler(cfg, authAdapter{module: users}, response.NewValidator())
	if err != nil {
		return nil, err
	}

	secHandler, err := securityhandler.NewSecurityHandler(cfg, securityAdapter{module: securities}, response.NewValidator())
	if err != nil {
		return nil, err
	}

	rabbitClient, err := brokerrabbit.NewClientFromConfig(cfg.RabbitMQ)
	if err != nil {
		return nil, err
	}

	ordHandler, err := ordershandler.NewOrdersHandler(cfg, ordersAdapter{module: orders}, response.NewValidator(), rabbitClient, httpServer.ShutdownSignal())
	if err != nil {
		rabbitClient.Close()
		return nil, err
	}

	modules := []module.Module{
		authHandler,
		secHandler,
		ordHandler,
	}
	for _, mod := range modules {
		httpServer.MountModule(apiPrefix, mod)
	}

	return &Server{http: httpServer, rabbit: rabbitClient}, nil
}

func (s *Server) Run(ctx context.Context) error {
	return s.http.Run(ctx)
}

// Start serves HTTP until the server is shut down.
func (s *Server) Start() error {
	return s.http.Start()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// Close releases the resources owned by the gateway.
func (s *Server) Close() error {
	return s.rabbit.Close()
}
