package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	activeorders "github.com/victorinolavida/orders/internal/application/usecases/active_orders"
	cancelorder "github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	createorder "github.com/victorinolavida/orders/internal/application/usecases/create_order"
	"github.com/victorinolavida/orders/internal/domain"
	orderpb "github.com/victorinolavida/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCHandler struct {
	orderpb.UnimplementedOrderServiceServer
	createOrder  *createorder.Usecase
	cancelOrder  *cancelorder.Usecase
	activeorders *activeorders.Usecase
}

func NewOrderGRPCHandler(
	createOrder *createorder.Usecase,
	cancelOrder *cancelorder.Usecase,
	activeorders *activeorders.Usecase,
) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		createOrder:  createOrder,
		cancelOrder:  cancelOrder,
		activeorders: activeorders,
	}
}

func parseScaled(s string) (int64, error) {
	if s == "" {
		return 0, nil // absent
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return 0, err
	}
	return d.Mul(decimal.NewFromInt(int64(domain.DecimalScale))).IntPart(), nil
}

func (h *OrderGRPCHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
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

	price, err := parseScaled(req.Price)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid price")
	}
	yieldVal, err := parseScaled(req.Yield)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid yield")
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
		Quantity:   req.Quantity,
	}

	res, err := h.createOrder.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orderpb.OrderResponse{
		Id:                res.ID,
		SecurityId:        res.SecurityId,
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
		CreatedAt:         timestamppb.New(res.CreatedAt),
		UpdatedAt:         timestamppb.New(res.UpdatedAt),
	}, nil
}

func (h *OrderGRPCHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.OrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	traderID, err := uuid.Parse(req.TraderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trader_id")
	}
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}

	cmd := cancelorder.Command{
		OrderID:  orderID,
		TraderID: traderID,
	}

	res, err := h.cancelOrder.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orderpb.OrderResponse{
		Id:                res.ID,
		SecurityId:        res.SecurityId,
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
		CreatedAt:         timestamppb.New(res.CreatedAt),
		UpdatedAt:         timestamppb.New(res.UpdatedAt),
	}, nil
}

func (h *OrderGRPCHandler) GetActiveOrders(ctx context.Context, req *orderpb.GetActiveOrdersRequest) (*orderpb.GetActiveOrdersResponse, error) {
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
	if req.Direction != "" {
		_, err := domain.NewDirection(req.Direction)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid direction")
		}
	}

	cmd := activeorders.Command{
		SecurityId: securityID,
		TraderID:   traderID,
		Direction:  req.Direction,
	}

	res, err := h.activeorders.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orders := make([]*orderpb.OrderResponse, len(res.Orders))
	for i, order := range res.Orders {
		ord := &orderpb.OrderResponse{
			Id:                order.ID,
			SecurityId:        order.SecurityId,
			Direction:         order.Direction,
			OrderType:         order.OrderType,
			Price:             order.Price,
			Yield:             order.Yield,
			Quantity:          order.Quantity,
			NominalValue:      order.NominalValue,
			AllocatedQuantity: order.AllocatedQuantity,
			AllocationPtc:     order.AllocationPtc,
			AllocationStatus:  order.AllocationStatus,
			Fx:                order.Fx,
			Status:            order.Status,
			CreatedAt:         timestamppb.New(order.CreatedAt),
			UpdatedAt:         timestamppb.New(order.UpdatedAt),
		}
		orders[i] = ord
	}

	return &orderpb.GetActiveOrdersResponse{
		Orders: orders,
	}, nil
}
