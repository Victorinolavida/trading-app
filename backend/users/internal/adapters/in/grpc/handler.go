package grpc

import (
	"context"
	"log/slog"

	"github.com/victorinolavida/auth/internal/application/usecases/login"
	"github.com/victorinolavida/auth/internal/application/usecases/register"
	authpb "github.com/victorinolavida/proto/auth"
)

type AuthGRPCHandler struct {
	authpb.UnimplementedAuthServiceServer
	registerHandler *register.UserRegisterHandler
	loginHandler    *login.UserLoginHandler
}

func NewAuthGRPCHandler(
	registerHandler *register.UserRegisterHandler,
	loginHandler *login.UserLoginHandler,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{registerHandler: registerHandler, loginHandler: loginHandler}
}

func (handler *AuthGRPCHandler) RegisterUser(ctx context.Context, req *authpb.RegisterRequest) (*authpb.AuthUserResponse, error) {
	result, err := handler.registerHandler.Handle(ctx, register.RegisterCommand{
		Email:     req.Email,
		Password:  req.Password,
		LastName:  req.LastName,
		FirstName: req.FirstName,
	})
	if err != nil {
		slog.Error("error registing user", slog.Any("error", err))
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
	result, err := handler.loginHandler.Handle(ctx, login.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		slog.Error("error registing user", slog.Any("error", err))
		return nil, err
	}

	return &authpb.AuthUserResponse{
		Id:        result.User.ID,
		FirstName: result.User.FirstName,
		LastName:  result.User.LastName,
		Token:     result.Token,
	}, nil
}
