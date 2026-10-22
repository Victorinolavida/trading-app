package ports

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
)

type AuthService interface {
	Login(context.Context, dtos.LoginRequest) (dtos.AuthUserResponse, error)
	Register(context.Context, dtos.RegisterRequest) (dtos.AuthUserResponse, error)
	Whoami(context.Context, dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error)
}
