package grpc

import (
	"context"
	"strings"

	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/register"
	"github.com/victorinolavida/auth/internal/application/usecases/whoami"
	authpb "github.com/victorinolavida/proto/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGRPCHandler struct {
	authpb.UnimplementedAuthServiceServer
	register *register.Usecase
	login    *login.Usecase
	whoami   *whoami.Usecase
}

func NewAuthGRPCHandler(
	registerUC *register.Usecase,
	loginUC *login.Usecase,
	whoamiUC *whoami.Usecase,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{register: registerUC, login: loginUC, whoami: whoamiUC}
}

func (handler *AuthGRPCHandler) RegisterUser(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthUserResponse, error) {
	command := register.Command{
		Email:     strings.TrimSpace(req.Email),
		Password:  strings.TrimSpace(req.Password),
		LastName:  strings.TrimSpace(req.LastName),
		FirstName: strings.TrimSpace(req.FirstName),
	}
	if command.Password == "" || command.Email == "" || command.LastName == "" || command.FirstName == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	result, err := handler.register.Execute(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authpb.AuthUserResponse{
		Id:        result.User.ID,
		FirstName: result.User.FirstName,
		LastName:  result.User.LastName,
		Token:     result.Token,
	}, nil
}

func (handler *AuthGRPCHandler) LoginUser(ctx context.Context, req *authpb.LoginRequest) (*authpb.AuthUserResponse, error) {
	command := login.Command{
		Email:    strings.TrimSpace(req.Email),
		Password: strings.TrimSpace(req.Password),
	}

	if command.Password == "" || command.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}
	result, err := handler.login.Execute(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authpb.AuthUserResponse{
		Id:        result.User.ID,
		FirstName: result.User.FirstName,
		LastName:  result.User.LastName,
		Token:     result.Token,
	}, nil
}

func (handler *AuthGRPCHandler) Whoami(ctx context.Context, req *authpb.WhoamiRequest) (*authpb.WhoamiResponse, error) {
	command := whoami.Command{
		UserID: strings.TrimSpace(req.UserId),
	}
	if command.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields")
	}

	result, err := handler.whoami.Execute(ctx, command)
	if err != nil {
		return nil, err
	}

	return &authpb.WhoamiResponse{
		Id:        result.UserID,
		Email:     result.Email,
		FirstName: result.FirstName,
		LastName:  result.LastName,
	}, nil
}
