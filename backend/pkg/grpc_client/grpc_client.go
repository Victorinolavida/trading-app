package grpcclient

import (
	"strconv"

	"github.com/victorinolavida/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultHost = "localhost"
	defaultPort = 8080
)

type GRPCClient struct {
	Client *grpc.ClientConn
}

func NewGRPCClient(conf config.GRPCClient) (*GRPCClient, error) {
	grpcClient, err := grpc.NewClient(formatGRPCAddress(conf.Host, conf.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCClient{Client: grpcClient}, nil
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

func (g *GRPCClient) Close() error {
	return g.Client.Close()
}
