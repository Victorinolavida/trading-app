package securityservice

import (
	"context"

	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	securitypb "github.com/victorinolavida/proto/security"
	pkgconfig "github.com/victorinolavida/pkg/config"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type SecurityService struct {
	client securitypb.AuthServiceClient
	grpc   *grpcclient.GRPCClient
}

var _ ports.SecurityService = (*SecurityService)(nil)

func NewSecurityService(conf pkgconfig.GRPCClient) (*SecurityService, error) {
	grpcClient, err := grpcclient.NewGRPCClient(conf)
	if err != nil {
		return nil, err
	}

	securityClient := securitypb.NewAuthServiceClient(grpcClient.Client)

	return &SecurityService{
		grpc:   grpcClient,
		client: securityClient,
	}, nil
}

func (serv *SecurityService) Close() error {
	return serv.grpc.Close()
}

func (serv *SecurityService) GetSecuritiesByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.ListSecuritiesResponse, error) {
	pbReq := &securitypb.ListSecuritiesRequest{
		Id:           req.Id,
		Isin:         req.Isin,
		Name:         req.Name,
		Q:            req.Query,
		Currency:     req.Currency,
		Serie:        req.Serie,
		SecurityType: req.SecurityType,
	}
	if pbReq.Q == "" && req.Q != "" {
		pbReq.Q = req.Q
	}

	if len(req.Ids) > 0 {
		// proto may not have repeated, but try to set if fields exist by converting - not possible; pass first
		pbReq.Id = req.Ids[0]
	}
	if len(req.Isins) > 0 {
		pbReq.Isin = req.Isins[0]
	}
	if len(req.Series) > 0 {
		pbReq.Serie = req.Series[0]
	}
	if len(req.SecurityTypes) > 0 {
		pbReq.SecurityType = req.SecurityTypes[0]
	}

	if req.IssueAfter != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueAfter); err == nil {
			pbReq.IssueAfter = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.IssueAfter); err == nil {
			pbReq.IssueAfter = timestamppb.New(t)
		}
	}
	if req.IssueBefore != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueBefore); err == nil {
			pbReq.IssueBefore = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.IssueBefore); err == nil {
			pbReq.IssueBefore = timestamppb.New(t)
		}
	}
	if req.MaturityAfter != "" {
		if t, err := time.Parse(time.RFC3339, req.MaturityAfter); err == nil {
			pbReq.MaturityAfter = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.MaturityAfter); err == nil {
			pbReq.MaturityAfter = timestamppb.New(t)
		}
	}
	if req.MaturityBefore != "" {
		if t, err := time.Parse(time.RFC3339, req.MaturityBefore); err == nil {
			pbReq.MaturityBefore = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.MaturityBefore); err == nil {
			pbReq.MaturityBefore = timestamppb.New(t)
		}
	}

	result, err := serv.client.GetSecuritiesByParams(ctx, pbReq)
	if err != nil {
		return dtos.ListSecuritiesResponse{}, err
	}

	resp := dtos.ListSecuritiesResponse{
		Securities: make([]dtos.Security, 0, len(result.Securities)),
	}
	for _, s := range result.Securities {
		var issue, maturity string
		if s.IssueDate != nil {
			issue = s.IssueDate.AsTime().Format("2006-01-02")
		}
		if s.MaturityDate != nil {
			maturity = s.MaturityDate.AsTime().Format("2006-01-02")
		}
		resp.Securities = append(resp.Securities, dtos.Security{
			ID:           s.Id,
			Isin:         s.Isin,
			Name:         s.Name,
			SecurityType: s.SecurityType,
			Serie:        s.Serie,
			IssueDate:    issue,
			MaturityDate: maturity,
			NominalValue: s.NominalValue,
			Currency:     s.Currency,
			Fx:           s.Fx,
			Enable:       s.Enable,
		})
	}
	return resp, nil
}

func (serv *SecurityService) GetByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.Security, error) {
	pbReq := &securitypb.ListSecuritiesRequest{
		Id:           req.Id,
		Isin:         req.Isin,
		Name:         req.Name,
		Q:            req.Query,
		Currency:     req.Currency,
		Serie:        req.Serie,
		SecurityType: req.SecurityType,
	}
	if pbReq.Q == "" && req.Q != "" {
		pbReq.Q = req.Q
	}
	if len(req.Ids) > 0 {
		pbReq.Id = req.Ids[0]
	}
	if len(req.Isins) > 0 {
		pbReq.Isin = req.Isins[0]
	}
	if len(req.Series) > 0 {
		pbReq.Serie = req.Series[0]
	}
	if len(req.SecurityTypes) > 0 {
		pbReq.SecurityType = req.SecurityTypes[0]
	}
	if req.IssueAfter != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueAfter); err == nil {
			pbReq.IssueAfter = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.IssueAfter); err == nil {
			pbReq.IssueAfter = timestamppb.New(t)
		}
	}
	if req.IssueBefore != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueBefore); err == nil {
			pbReq.IssueBefore = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.IssueBefore); err == nil {
			pbReq.IssueBefore = timestamppb.New(t)
		}
	}
	if req.MaturityAfter != "" {
		if t, err := time.Parse(time.RFC3339, req.MaturityAfter); err == nil {
			pbReq.MaturityAfter = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.MaturityAfter); err == nil {
			pbReq.MaturityAfter = timestamppb.New(t)
		}
	}
	if req.MaturityBefore != "" {
		if t, err := time.Parse(time.RFC3339, req.MaturityBefore); err == nil {
			pbReq.MaturityBefore = timestamppb.New(t)
		} else if t, err := time.Parse("2006-01-02", req.MaturityBefore); err == nil {
			pbReq.MaturityBefore = timestamppb.New(t)
		}
	}

	s, err := serv.client.GetByParams(ctx, pbReq)
	if err != nil {
		return dtos.Security{}, err
	}

	var issue, maturity string
	if s.IssueDate != nil {
		issue = s.IssueDate.AsTime().Format("2006-01-02")
	}
	if s.MaturityDate != nil {
		maturity = s.MaturityDate.AsTime().Format("2006-01-02")
	}

	return dtos.Security{
		ID:           s.Id,
		Isin:         s.Isin,
		Name:         s.Name,
		SecurityType: s.SecurityType,
		Serie:        s.Serie,
		IssueDate:    issue,
		MaturityDate: maturity,
		NominalValue: s.NominalValue,
		Currency:     s.Currency,
		Fx:           s.Fx,
		Enable:       s.Enable,
	}, nil
}

func (serv *SecurityService) GetEnableSecurities(ctx context.Context) (dtos.ListSecuritiesResponse, error) {
	result, err := serv.client.GetEnableSecurities(ctx, &securitypb.Empty{})
	if err != nil {
		return dtos.ListSecuritiesResponse{}, err
	}

	resp := dtos.ListSecuritiesResponse{
		Securities: make([]dtos.Security, 0, len(result.Securities)),
	}
	for _, s := range result.Securities {
		var issue, maturity string
		if s.IssueDate != nil {
			issue = s.IssueDate.AsTime().Format("2006-01-02")
		}
		if s.MaturityDate != nil {
			maturity = s.MaturityDate.AsTime().Format("2006-01-02")
		}
		resp.Securities = append(resp.Securities, dtos.Security{
			ID:           s.Id,
			Isin:         s.Isin,
			Name:         s.Name,
			SecurityType: s.SecurityType,
			Serie:        s.Serie,
			IssueDate:    issue,
			MaturityDate: maturity,
			NominalValue: s.NominalValue,
			Currency:     s.Currency,
			Fx:           s.Fx,
			Enable:       s.Enable,
		})
	}
	return resp, nil
}

