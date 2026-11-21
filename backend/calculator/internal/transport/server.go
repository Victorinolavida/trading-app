package transport

import (
	"context"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/calculator/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

type Server struct{ Module *api.Module }

func input(req *structpb.Struct, requireDate bool) (string, time.Time, error) {
	if req == nil {
		return "", time.Time{}, status.Error(codes.InvalidArgument, "request is required")
	}
	id := req.GetFields()["security_id"].GetStringValue()
	if _, err := uuid.Parse(id); err != nil {
		return "", time.Time{}, status.Error(codes.InvalidArgument, "invalid security_id")
	}
	value := req.GetFields()["settlement_date"].GetStringValue()
	if value == "" && !requireDate {
		return id, time.Time{}, nil
	}
	date, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", time.Time{}, status.Error(codes.InvalidArgument, "invalid settlement_date")
	}
	return id, date, nil
}
func numeric(req *structpb.Struct, key string) (string, error) {
	value := req.GetFields()[key].GetStringValue()
	if _, err := decimal.NewFromString(value); err != nil {
		return "", status.Errorf(codes.InvalidArgument, "invalid %s", key)
	}
	return value, nil
}
func rpcError(err error) error {
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.Unknown {
		return err
	}
	return status.Error(codes.InvalidArgument, err.Error())
}
func (s *Server) Price(ctx context.Context, r *structpb.Struct) (*structpb.Struct, error) {
	id, date, err := input(r, true)
	if err != nil {
		return nil, err
	}
	y, err := numeric(r, "yield")
	if err != nil {
		return nil, err
	}
	result, err := s.Module.Price(ctx, api.PriceRequest{SecurityID: id, SettlementDate: date, Yield: y})
	if err != nil {
		return nil, rpcError(err)
	}
	return structpb.NewStruct(map[string]any{"clean_price": result.CleanPrice, "accrued_interest": result.AccruedInterest, "dirty_price": result.DirtyPrice, "mxn_value": result.MXNValue})
}
func (s *Server) Yield(ctx context.Context, r *structpb.Struct) (*structpb.Struct, error) {
	id, date, err := input(r, true)
	if err != nil {
		return nil, err
	}
	price, err := numeric(r, "clean_price")
	if err != nil {
		return nil, err
	}
	result, err := s.Module.Yield(ctx, api.YieldRequest{SecurityID: id, SettlementDate: date, CleanPrice: price})
	if err != nil {
		return nil, rpcError(err)
	}
	return structpb.NewStruct(map[string]any{"yield": result})
}
func (s *Server) Cashflows(ctx context.Context, r *structpb.Struct) (*structpb.Struct, error) {
	id, date, err := input(r, false)
	if err != nil {
		return nil, err
	}
	result, err := s.Module.Cashflows(ctx, api.CashflowRequest{SecurityID: id, SettlementDate: date})
	if err != nil {
		return nil, rpcError(err)
	}
	flows := make([]any, 0, len(result))
	for _, cf := range result {
		flows = append(flows, map[string]any{"date": cf.Date.Format(time.RFC3339), "rate": cf.Amount.String()})
	}
	return structpb.NewStruct(map[string]any{"cashflows": flows})
}
