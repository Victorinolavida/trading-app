package ports

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
)

type OrdersService interface {
	CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.OrderResponse, error)
	CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.OrderResponse, error)
	ListOrders(ctx context.Context, req dtos.ListOrdersRequest) (dtos.ListOrdersResponse, error)
	GetActiveOrders(ctx context.Context, req dtos.GetActiveOrdersRequest) (dtos.GetActiveOrdersResponse, error)
	Close() error
}
