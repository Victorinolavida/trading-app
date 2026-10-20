package grpc

import (
	"context"

	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/register"
	authpb "github.com/victorinolavida/proto/auth"
)

type AuthGRPCHandler struct {
	authpb.UnimplementedAuthServiceServer
	register *register.Usecase
	login    *login.Usecase
}

func NewAuthGRPCHandler(
	registerUC *register.Usecase,
	loginUC *login.Usecase,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{register: registerUC, login: loginUC}
}

func (handler *AuthGRPCHandler) RegisterUser(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthUserResponse, error) {
	result, err := handler.register.Execute(ctx, register.Command{
		Email:     req.Email,
		Password:  req.Password,
		LastName:  req.LastName,
		FirstName: req.FirstName,
	})
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
	result, err := handler.login.Execute(ctx, login.Command{
		Email:    req.Email,
		Password: req.Password,
	})
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
