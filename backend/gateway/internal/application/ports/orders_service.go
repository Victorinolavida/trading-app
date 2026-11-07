package ports

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.OrderResponse, error)
	CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.OrderResponse, error)
	GetOrderByID(ctx context.Context, req dtos.GetOrderRequest) (dtos.OrderResponse, error)
	GetActiveOrders(ctx context.Context, req dtos.GetActiveOrdersRequest) (dtos.GetActiveOrdersResponse, error)
	Close() error
}
