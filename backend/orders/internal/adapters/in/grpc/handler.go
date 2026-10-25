package grpc

import (
	"context"

	"github.com/google/uuid"
	cancelorder "github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	createorder "github.com/victorinolavida/orders/internal/application/usecases/create_order"
	getordergroup "github.com/victorinolavida/orders/internal/application/usecases/get_order_group"
	orderpb "github.com/victorinolavida/proto/order"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCHandler struct {
	orderpb.UnimplementedOrderServiceServer
	createOrder    *createorder.Usecase
	cancelOrder    *cancelorder.Usecase
	getOrderGroup  *getordergroup.Usecase
}

func NewOrderGRPCHandler(
	createOrder *createorder.Usecase,
	cancelOrder *cancelorder.Usecase,
	getOrderGroup *getordergroup.Usecase,
) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		createOrder:   createOrder,
		cancelOrder:   cancelOrder,
		getOrderGroup: getOrderGroup,
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
	qty, err := decimal.NewFromString(req.Quantity)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid quantity")
	}

	cmd := createorder.Command{
		TraderID:   traderID,
		SecurityID: securityID,
		Direction:  req.Direction,
		OrderType:  req.OrderType,
		Price:      price,
		Yield:      yieldVal,
		Quantity:   qty,
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

func (h *OrderGRPCHandler) GetOrderGroup(ctx context.Context, req *orderpb.GetOrderGroupRequest) (*orderpb.GetOrderGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	orderGroupID, err := uuid.Parse(req.OrderGroupId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_group_id")
	}

	res, err := h.getOrderGroup.Execute(ctx, orderGroupID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &orderpb.GetOrderGroupResponse{
		Id:               res.ID,
		PublicId:         res.PublicID,
		TraderId:         res.TraderID,
		SecurityId:       res.SecurityID,
		Direction:        res.Direction,
		OrderType:        res.OrderType,
		AllocationPc:     res.AllocationPC.String(),
		AllocationStatus: res.AllocationStatus,
		OrderGroupStatus: res.OrderGroupStatus,
		WeightedPrice:    res.WeightedPrice.String(),
		WeightedYield:    res.WeightedYield.String(),
		WeightedFx:       res.WeightedFX.String(),
	}

	for _, o := range res.Orders {
		resp.Orders = append(resp.Orders, &orderpb.Order{
			Id:           o.ID,
			OrderGroupId: o.OrderGroupID,
			TraderId:     o.TraderID,
			SecurityId:   o.SecurityID,
			Direction:    o.Direction,
			OrderType:    o.OrderType,
			Price:        o.Price.String(),
			Yield:        o.Yield.String(),
			Quantity:     o.Quantity.String(),
			NominalValue: o.NominalValue.String(),
			Fx:           o.Fx.String(),
			Status:       o.Status,
		})
	}

	return resp, nil
}
