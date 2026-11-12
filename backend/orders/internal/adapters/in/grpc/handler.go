package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	activeorders "github.com/victorinolavida/orders/internal/application/usecases/active_orders"
	cancelorder "github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	createorder "github.com/victorinolavida/orders/internal/application/usecases/create_order"
	getmarketdepth "github.com/victorinolavida/orders/internal/application/usecases/get_market_depth"
	listorders "github.com/victorinolavida/orders/internal/application/usecases/list_orders"
	"github.com/victorinolavida/orders/internal/domain"
	orderpb "github.com/victorinolavida/proto/order"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGRPCHandler struct {
	orderpb.UnimplementedOrderServiceServer
	createOrder    *createorder.Usecase
	cancelOrder    *cancelorder.Usecase
	activeorders   *activeorders.Usecase
	listorders     *listorders.Usecase
	getMarketDepth *getmarketdepth.Usecase
}

func NewOrderGRPCHandler(
	createOrder *createorder.Usecase,
	cancelOrder *cancelorder.Usecase,
	activeorders *activeorders.Usecase,
	listorders *listorders.Usecase,
	getMarketDepth *getmarketdepth.Usecase,
) *OrderGRPCHandler {
	return &OrderGRPCHandler{
		createOrder:    createOrder,
		cancelOrder:    cancelOrder,
		activeorders:   activeorders,
		listorders:     listorders,
		getMarketDepth: getMarketDepth,
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
		ExpiresAt:  req.ExpiresAt.AsDuration(),
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
		ExpiresAt:         timestamppb.New(res.ExpiresAt),
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
		ExpiresAt:         timestamppb.New(res.ExpiresAt),
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
			ExpiresAt:         timestamppb.New(order.ExpiresAt),
		}
		orders[i] = ord
	}

	return &orderpb.GetActiveOrdersResponse{
		Orders: orders,
	}, nil
}

func (h *OrderGRPCHandler) GetMarketDepth(ctx context.Context, req *orderpb.GetMarketDepthRequest) (*orderpb.GetMarketDepthResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	securityID, err := uuid.Parse(req.SecurityId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid security_id")
	}

	cmd := getmarketdepth.Command{
		SecurityID: securityID,
	}

	res, err := h.getMarketDepth.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if res.MarketDepth == nil {
		return &orderpb.GetMarketDepthResponse{}, nil
	}

	bids := make([]*orderpb.MarketDepthLevel, len(res.MarketDepth.Bid))
	for i, side := range res.MarketDepth.Bid {
		bids[i] = toProtoMarketDepthLevel(side)
	}

	asks := make([]*orderpb.MarketDepthLevel, len(res.MarketDepth.Ask))
	for i, side := range res.MarketDepth.Ask {
		asks[i] = toProtoMarketDepthLevel(side)
	}

	return &orderpb.GetMarketDepthResponse{
		MarketDepth: &orderpb.MarketDepth{
			Bids: bids,
			Asks: asks,
		},
	}, nil
}

func toProtoMarketDepthLevel(side *domain.MarketDepthSide) *orderpb.MarketDepthLevel {
	if side == nil {
		return nil
	}
	return &orderpb.MarketDepthLevel{
		Price:        side.Price.String(),
		NominalValue: side.AggNominalValue.String(),
		Quantity:     side.AggQuantity.Value(),
		OrderCounts:  side.OrderCounts,
		Direction:    side.Direction.String(),
	}
}

func (h *OrderGRPCHandler) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	traderID, err := uuid.Parse(req.TraderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid trader_id")
	}

	orderIDs := make([]uuid.UUID, 0, len(req.OrderIds))
	for _, id := range req.OrderIds {
		orderID, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid order_id")
		}
		orderIDs = append(orderIDs, orderID)
	}

	if req.Direction != "" {
		if _, err := domain.NewDirection(req.Direction); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid direction")
		}
	}
	if req.Status != "" {
		if _, err := domain.NewOrderStatus(req.Status); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid status")
		}
	}

	cmd := listorders.Command{
		TraderID:  traderID,
		OrderIDs:  orderIDs,
		Direction: req.Direction,
		Status:    req.Status,
	}

	res, err := h.listorders.Execute(ctx, cmd)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	orders := make([]*orderpb.OrderResponse, len(res.Orders))
	for i, order := range res.Orders {
		orders[i] = &orderpb.OrderResponse{
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
			ExpiresAt:         timestamppb.New(order.ExpiresAt),
			UpdatedAt:         timestamppb.New(order.UpdatedAt),
		}
	}

	return &orderpb.ListOrdersResponse{
		Orders: orders,
	}, nil
}
