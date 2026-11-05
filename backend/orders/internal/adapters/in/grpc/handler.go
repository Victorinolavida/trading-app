package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	cancelorder "github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	createorder "github.com/victorinolavida/orders/internal/application/usecases/create_order"
	orderpb "github.com/victorinolavida/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCHandler struct {
	orderpb.UnimplementedOrderServiceServer
	createOrder *createorder.Usecase
	cancelOrder *cancelorder.Usecase
}

func NewOrderGRPCHandler(
	createOrder *createorder.Usecase,
	cancelOrder *cancelorder.Usecase,
) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		createOrder: createOrder,
		cancelOrder: cancelOrder,
	}
}

func (h *OrderGRPCHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	traderID, err := uuid.Parse(req.TraderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trader_id")
	}
	securityID, err := uuid.Parse(req.SecurityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid security_id")
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		price = decimal.Zero
	}
	yieldVal, err := decimal.NewFromString(req.Yield)
	if err != nil {
		yieldVal = decimal.Zero
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid quantity")
	}

	cmd := createorder.Command{
		TraderID:   traderID,
		SecurityID: securityID,
		Direction:  req.Direction,
		OrderType:  req.OrderType,
		Price:      price,
		Yield:      yieldVal,
		Quantity:   uint(req.Quantity),
	}

	res, err := h.createOrder.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orderpb.CreateOrderResponse{
		OrderGroupId: res.OrderGroupID,
		OrderId:      res.OrderID,
		Status:       res.Status,
	}, nil
}

func (h *OrderGRPCHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	traderID, err := uuid.Parse(req.TraderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trader_id")
	}
	orderGroupID, err := uuid.Parse(req.OrderGroupId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_group_id")
	}

	cmd := cancelorder.Command{
		OrderGroupID: orderGroupID,
		TraderID:     traderID,
	}

	res, err := h.cancelOrder.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orderpb.CancelOrderResponse{
		OrderGroupId: res.OrderGroupID,
		Status:       res.Status,
	}, nil
}
