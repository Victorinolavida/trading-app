package grpc

import (
	"context"
	"log/slog"
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
	register      *register.Usecase
	login         *login.Usecase
	whoami        *whoami.Usecase
	passwordReset passwordResetUsecase
}

type passwordResetUsecase interface {
	Request(context.Context, string) error
	Confirm(context.Context, string, string) error
}

func NewAuthGRPCHandler(
	registerUC *register.Usecase,
	loginUC *login.Usecase,
	whoamiUC *whoami.Usecase,
	resetUC passwordResetUsecase,
) *AuthGRPCHandler {
	return &AuthGRPCHandler{register: registerUC, login: loginUC, whoami: whoamiUC, passwordReset: resetUC}
}

func (handler *AuthGRPCHandler) RequestPasswordReset(ctx context.Context, req *authpb.RequestPasswordResetRequest) (*authpb.PasswordResetResponse, error) {
	if handler.passwordReset == nil {
		return nil, status.Error(codes.Unimplemented, "password reset is not available")
	}
	// Do not reveal whether a user exists or whether delivery succeeded.
	if err := handler.passwordReset.Request(ctx, strings.TrimSpace(req.GetEmail())); err != nil {
		slog.Error("password reset request could not be processed")
	}
	return &authpb.PasswordResetResponse{}, nil
}

func (handler *AuthGRPCHandler) ConfirmPasswordReset(ctx context.Context, req *authpb.ConfirmPasswordResetRequest) (*authpb.PasswordResetResponse, error) {
	if handler.passwordReset == nil {
		return nil, status.Error(codes.Unimplemented, "password reset is not available")
	}
	if err := handler.passwordReset.Confirm(ctx, req.GetToken(), req.GetNewPassword()); err != nil {
		return nil, err
	}
	return &authpb.PasswordResetResponse{}, nil
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
		Id:             result.UserID,
		Email:          result.Email,
		FirstName:      result.FirstName,
		LastName:       result.LastName,
		SessionVersion: result.SessionVersion,
	}, nil
}
