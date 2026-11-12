package ordersservice

import (
	"context"
	"time"

	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	pkgconfig "github.com/victorinolavida/pkg/config"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	orderpb "github.com/victorinolavida/proto/order"
	"google.golang.org/protobuf/types/known/durationpb"
)

type OrdersService struct {
	client orderpb.OrderServiceClient
	grpc   *grpcclient.GRPCClient
}

var _ ports.OrdersService = (*OrdersService)(nil)

func NewOrdersService(conf pkgconfig.GRPCClient) (*OrdersService, error) {
	grpcClient, err := grpcclient.NewGRPCClient(conf)
	if err != nil {
		return nil, err
	}

	client := orderpb.NewOrderServiceClient(grpcClient.Client)

	return &OrdersService{
		grpc:   grpcClient,
		client: client,
	}, nil
}

func (s *OrdersService) Close() error {
	return s.grpc.Close()
}

func toOrderResponse(res *orderpb.OrderResponse) dtos.OrderResponse {
	if res == nil {
		return dtos.OrderResponse{}
	}
	var createdAt, updatedAt, expiresAt time.Time
	if res.CreatedAt != nil {
		createdAt = res.CreatedAt.AsTime()
	}
	if res.UpdatedAt != nil {
		updatedAt = res.UpdatedAt.AsTime()
	}
	if res.ExpiresAt != nil {
		expiresAt = res.ExpiresAt.AsTime()
	}
	return dtos.OrderResponse{
		ID:                res.Id,
		SecurityID:        res.SecurityId,
		Direction:         res.Direction,
		OrderType:         res.OrderType,
		Price:             res.Price,
		Yield:             res.Yield,
		Quantity:          res.Quantity,
		NominalValue:      res.NominalValue,
		AllocatedQuantity: res.AllocatedQuantity,
		AllocationPtc:     res.AllocationPtc,
		AllocationStatus:  res.AllocationStatus,
		Fx:                res.Fx,
		Status:            res.Status,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		ExpiresAt:         expiresAt,
	}
}

func (s *OrdersService) CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.OrderResponse, error) {
	res, err := s.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
		TraderId:   req.TraderID,
		SecurityId: req.SecurityID,
		Direction:  req.Direction,
		OrderType:  req.OrderType,
		Price:      req.Price,
		Yield:      req.Yield,
		Quantity:   req.Quantity,
		ExpiresAt:  durationpb.New(req.Expiration.Value()),
	})
	if err != nil {
		return dtos.OrderResponse{}, err
	}
	return toOrderResponse(res), nil
}

func (s *OrdersService) CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.OrderResponse, error) {
	res, err := s.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{
		OrderId:  req.OrderID,
		TraderId: req.TraderID,
	})
	if err != nil {
		return dtos.OrderResponse{}, err
	}
	return toOrderResponse(res), nil
}

func (s *OrdersService) ListOrders(ctx context.Context, req dtos.ListOrdersRequest) (dtos.ListOrdersResponse, error) {
	res, err := s.client.ListOrders(ctx, &orderpb.ListOrdersRequest{
		TraderId:  req.TraderID,
		OrderIds:  req.OrderIDs,
		Direction: req.Direction,
		Status:    req.Status,
	})
	if err != nil {
		return dtos.ListOrdersResponse{}, err
	}
	orders := make([]dtos.OrderResponse, 0, len(res.Orders))
	for _, o := range res.Orders {
		orders = append(orders, toOrderResponse(o))
	}
	return dtos.ListOrdersResponse{Orders: orders}, nil
}

func (s *OrdersService) GetActiveOrders(ctx context.Context, req dtos.GetActiveOrdersRequest) (dtos.GetActiveOrdersResponse, error) {
	res, err := s.client.GetActiveOrders(ctx, &orderpb.GetActiveOrdersRequest{
		SecurityId: req.SecurityID,
		Direction:  req.Direction,
		TraderId:   req.TraderID,
	})
	if err != nil {
		return dtos.GetActiveOrdersResponse{}, err
	}
	orders := make([]dtos.OrderResponse, 0, len(res.Orders))
	for _, o := range res.Orders {
		orders = append(orders, toOrderResponse(o))
	}
	return dtos.GetActiveOrdersResponse{Orders: orders}, nil
}

func (s *OrdersService) GetMarketDepth(ctx context.Context, req dtos.GetMarketDepthRequest) (dtos.MarketDepthResponse, error) {
	res, err := s.client.GetMarketDepth(ctx, &orderpb.GetMarketDepthRequest{
		SecurityId: req.SecurityID,
	})
	if err != nil {
		return dtos.MarketDepthResponse{}, err
	}

	depth := dtos.MarketDepthResponse{}
	if res.GetMarketDepth() != nil {
		depth.Bids = make([]*dtos.MarketDepthLevel, 0, len(res.GetMarketDepth().GetBids()))
		for _, bid := range res.GetMarketDepth().GetBids() {
			if bid != nil {
				depth.Bids = append(depth.Bids, toMarketDepthLevel(bid))
			}
		}
		depth.Asks = make([]*dtos.MarketDepthLevel, 0, len(res.GetMarketDepth().GetAsks()))
		for _, ask := range res.GetMarketDepth().GetAsks() {
			if ask != nil {
				depth.Asks = append(depth.Asks, toMarketDepthLevel(ask))
			}
		}
	}

	return depth, nil
}

func toMarketDepthLevel(level *orderpb.MarketDepthLevel) *dtos.MarketDepthLevel {
	if level == nil {
		return nil
	}
	return &dtos.MarketDepthLevel{
		Price:        level.GetPrice(),
		NominalValue: level.GetNominalValue(),
		Quantity:     level.GetQuantity(),
		OrderCounts:  level.GetOrderCounts(),
		Direction:    level.GetDirection(),
	}
}
