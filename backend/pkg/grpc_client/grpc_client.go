package grpcclient

import (
	"context"
	"strconv"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/victorinolavida/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(RequestMetaInterceptor),
	)
	if err != nil {
		return nil, err
	}
	return &GRPCClient{Client: grpcClient}, nil
}

// RequestMetaInterceptor forwards the request ID and gateway request metadata.
func RequestMetaInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	reqID := middleware.GetReqID(ctx)
	if reqID == "" {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-request-id"); len(ids) > 0 {
				reqID = ids[0]
			}
		}
	}
	if reqID != "" {
		md, _ := metadata.FromOutgoingContext(ctx)
		md = md.Copy()
		md.Set("x-request-id", reqID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	if ip, ok := ctx.Value(clientIPKey).(string); ok && ip != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-client-ip", ip)
	}
	if ua, ok := ctx.Value(userAgentKey).(string); ok && ua != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user-agent", ua)
	}
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

type ctxKey string

const (
	clientIPKey  ctxKey = "client_ip"
	userAgentKey ctxKey = "user_agent"
	userIDKey    ctxKey = "user_id"
)

func WithClientIP(ctx context.Context, ip string) context.Context {
	if ip == "" {
		return ctx
	}
	return context.WithValue(ctx, clientIPKey, ip)
}

func WithUserAgent(ctx context.Context, ua string) context.Context {
	if ua == "" {
		return ctx
	}
	return context.WithValue(ctx, userAgentKey, ua)
}

func WithUserID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, userIDKey, id)
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
