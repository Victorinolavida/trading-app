package ports

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.CreateOrderResponse, error)
	CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.CancelOrderResponse, error)
	GetOrderGroup(ctx context.Context, orderGroupID string) (dtos.OrderGroupDetail, error)
	Close() error
}
