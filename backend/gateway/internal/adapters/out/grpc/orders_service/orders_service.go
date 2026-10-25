package ordersservice

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	orderpb "github.com/victorinolavida/proto/order"
	pkgconfig "github.com/victorinolavida/pkg/config"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
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

func (s *OrdersService) CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.CreateOrderResponse, error) {
	res, err := s.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
		TraderId:   req.TraderID,
		SecurityId: req.SecurityID,
		Direction:  req.Direction,
		OrderType:  req.OrderType,
		Price:      req.Price,
		Yield:      req.Yield,
		Quantity:   req.Quantity,
	})
	if err != nil {
		return dtos.CreateOrderResponse{}, err
	}
	return dtos.CreateOrderResponse{
		OrderGroupID: res.OrderGroupId,
		OrderID:      res.OrderId,
		Status:       res.Status,
	}, nil
}

func (s *OrdersService) CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.CancelOrderResponse, error) {
	res, err := s.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{
		OrderGroupId: req.OrderGroupID,
		TraderId:     req.TraderID,
	})
	if err != nil {
		return dtos.CancelOrderResponse{}, err
	}
	return dtos.CancelOrderResponse{
		OrderGroupID: res.OrderGroupId,
		Status:       res.Status,
	}, nil
}

func (s *OrdersService) GetOrderGroup(ctx context.Context, orderGroupID string) (dtos.OrderGroupDetail, error) {
	res, err := s.client.GetOrderGroup(ctx, &orderpb.GetOrderGroupRequest{
		OrderGroupId: orderGroupID,
	})
	if err != nil {
		return dtos.OrderGroupDetail{}, err
	}

	orders := make([]dtos.Order, 0, len(res.Orders))
	for _, o := range res.Orders {
		orders = append(orders, dtos.Order{
			ID:           o.Id,
			OrderGroupID: o.OrderGroupId,
			TraderID:     o.TraderId,
			SecurityID:   o.SecurityId,
			Direction:    o.Direction,
			OrderType:    o.OrderType,
			Price:        o.Price,
			Yield:        o.Yield,
			Quantity:     o.Quantity,
			NominalValue: o.NominalValue,
			Fx:           o.Fx,
			Status:       o.Status,
		})
	}

	return dtos.OrderGroupDetail{
		ID:               res.Id,
		PublicID:         res.PublicId,
		TraderID:         res.TraderId,
		SecurityID:       res.SecurityId,
		Direction:        res.Direction,
		OrderType:        res.OrderType,
		AllocationPC:     res.AllocationPc,
		AllocationStatus: res.AllocationStatus,
		OrderGroupStatus: res.OrderGroupStatus,
		WeightedPrice:    res.WeightedPrice,
		WeightedYield:    res.WeightedYield,
		WeightedFX:       res.WeightedFx,
		Orders:           orders,
	}, nil
}
