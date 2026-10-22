package authservice

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/config"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	authpb "github.com/victorinolavida/proto/auth"
)

type AuthService struct {
	client authpb.AuthServiceClient
	conn   *grpcclient.GRPCClient
}

var _ ports.AuthService = (*AuthService)(nil)

func withRequestMeta(ctx context.Context, r *http.Request) context.Context {
	ctx = grpcclient.WithClientIP(ctx, middleware.GetClientIP(r.Context()))
	ctx = grpcclient.WithUserAgent(ctx, r.UserAgent())
	return ctx
}

func NewAuthService(conf config.GRPCClient) (*AuthService, error) {
	grpcClient, err := grpcclient.NewGRPCClient(conf)
	if err != nil {
		return nil, err
	}
	authClient := authpb.NewAuthServiceClient(grpcClient.Client)
	return &AuthService{
		client: authClient,
		conn:   grpcClient,
	}, nil
}

// Close releases the gRPC connection held by the adapter.
func (serv *AuthService) Close() error {
	return serv.conn.Close()
}

func (serv *AuthService) Login(ctx context.Context, request dtos.LoginRequest) (dtos.AuthUserResponse, error) {
	result, err := serv.client.LoginUser(ctx, &authpb.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return dtos.AuthUserResponse{}, err
	}
	return dtos.AuthUserResponse{
		UserID:    result.Id,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Token:     result.Token,
	}, nil
}

func (serv *AuthService) Register(ctx context.Context, command dtos.RegisterRequest) (dtos.AuthUserResponse, error) {
	result, err := serv.client.RegisterUser(ctx, &authpb.RegisterRequest{
		Email:     command.Email,
		FirstName: command.FirstName,
		LastName:  command.LastName,
		Password:  command.Password,
	})
	if err != nil {
		return dtos.AuthUserResponse{}, err
	}
	return dtos.AuthUserResponse{
		UserID:    result.Id,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Token:     result.Token,
	}, nil
}

func (serv *AuthService) Whoami(ctx context.Context, request dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error) {
	result, err := serv.client.Whoami(ctx, &authpb.WhoamiRequest{
		Token: request.Token,
	})
	if err != nil {
		return dtos.WhoamiUserResponse{}, err
	}
	return dtos.WhoamiUserResponse{
		UserID:    result.Id,
		FirstName: result.FirstName,
		LastName:  result.LastName,
		Email:     result.Email,
	}, nil
}
