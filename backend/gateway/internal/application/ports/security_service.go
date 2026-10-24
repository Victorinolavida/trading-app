package ports

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
)

type SecurityService interface {
	GetSecuritiesByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.ListSecuritiesResponse, error)
	GetByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.Security, error)
	GetEnableSecurities(ctx context.Context) (dtos.ListSecuritiesResponse, error)
	Close() error
}
