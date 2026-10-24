package grcp

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	getenablesecurities "github.com/victorinolavida/securities/internal/application/usecases/get_enable_securities"
	getsecurities "github.com/victorinolavida/securities/internal/application/usecases/get_securities"
	getsecuritybyparams "github.com/victorinolavida/securities/internal/application/usecases/get_security_by_params"
	"github.com/victorinolavida/securities/internal/application/ports"
	"github.com/victorinolavida/securities/internal/domain"
	securitypb "github.com/victorinolavida/proto/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SecurityGRPCHandler struct {
	securitypb.UnimplementedAuthServiceServer
	getEnable   *getenablesecurities.Usecase
	getByParams *getsecurities.Usecase
	getByParam  *getsecuritybyparams.Usecase
}

func NewSecurityGRPCHandler(
	getEnable *getenablesecurities.Usecase,
	getByParams *getsecurities.Usecase,
	getByParam *getsecuritybyparams.Usecase,
) *SecurityGRPCHandler {
	return &SecurityGRPCHandler{
		getEnable:   getEnable,
		getByParams: getByParams,
		getByParam:  getByParam,
	}
}

func (h *SecurityGRPCHandler) GetSecuritiesByParams(ctx context.Context, req *securitypb.ListSecuritiesRequest) (*securitypb.ListSecuritiesResponse, error) {
	filters := ports.Params{
		TimeFilters: ports.TimeFilters{},
	}

	if req != nil {
		if req.Id != "" {
			parsed, err := uuid.Parse(strings.TrimSpace(req.Id))
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid id")
			}
			filters.IDs = uuid.UUIDs{parsed}
		}

		if req.Isin != "" {
			filters.ISINs = []string{req.Isin}
		}

		if req.Serie != "" {
			filters.Serie = []string{req.Serie}
		}

		if req.SecurityType != "" {
			filters.SecurityTypes = []domain.SecurityType{domain.SecurityType(req.SecurityType)}
		}

		if req.Name != "" {
			filters.Name = req.Name
		}
		if req.Q != "" {
			filters.Query = req.Q
		}

		if req.IssueAfter != nil {
			filters.TimeFilters.IssueAfter = req.IssueAfter.AsTime()
		}
		if req.IssueBefore != nil {
			filters.TimeFilters.IssueBefore = req.IssueBefore.AsTime()
		}
		if req.MaturityAfter != nil {
			filters.TimeFilters.MaturityAfter = req.MaturityAfter.AsTime()
		}
		if req.MaturityBefore != nil {
			filters.TimeFilters.MaturityBefore = req.MaturityBefore.AsTime()
		}
	}

	res, err := h.getByParams.Execute(ctx, filters)
	if err != nil {
		return nil, err
	}

	securities := make([]*securitypb.Security, 0, len(res.Results))
	for _, s := range res.Results {
		issueDate, _ := time.Parse("2006-01-02", s.IssueDate)
		maturityDate, _ := time.Parse("2006-01-02", s.MaturityDate)
		securities = append(securities, &securitypb.Security{
			Id:           s.ID,
			Isin:         s.Isin,
			Name:         s.Name,
			SecurityType: s.Type,
			Serie:        s.Serie,
			IssueDate:    timestamppb.New(issueDate),
			MaturityDate: timestamppb.New(maturityDate),
			NominalValue: parseFloat(s.NominalValue),
			Currency:     s.Currency,
			Fx:           parseFloat(s.Fx),
			Enable:       s.Enable,
		})
	}

	return &securitypb.ListSecuritiesResponse{Securities: securities}, nil
}

func (h *SecurityGRPCHandler) GetByParams(ctx context.Context, req *securitypb.ListSecuritiesRequest) (*securitypb.Security, error) {
	cmd := getsecuritybyparams.Command{}
	if req != nil {
		cmd.ID = strings.TrimSpace(req.Id)
		cmd.ISIN = strings.TrimSpace(req.Isin)
		cmd.Serie = strings.TrimSpace(req.Serie)
		cmd.SecurityType = strings.TrimSpace(req.SecurityType)
		if req.Name != "" {
			cmd.ISIN = strings.TrimSpace(req.Name)
		}
		if req.Q != "" {
			cmd.ISIN = strings.TrimSpace(req.Q)
		}
		if req.IssueAfter != nil {
			cmd.IssueAfter = req.IssueAfter.AsTime()
		}
		if req.IssueBefore != nil {
			cmd.IssueBefore = req.IssueBefore.AsTime()
		}
		if req.MaturityAfter != nil {
			cmd.MaturityAfter = req.MaturityAfter.AsTime()
		}
		if req.MaturityBefore != nil {
			cmd.MaturityBefore = req.MaturityBefore.AsTime()
		}
	}

	res, err := h.getByParam.Execute(ctx, cmd)
	if err != nil {
		return nil, err
	}
	if res.Security.ID == "" && res.Security.Isin == "" {
		return nil, status.Error(codes.NotFound, "security not found")
	}

	issueDate, _ := time.Parse("2006-01-02", res.Security.IssueDate)
	maturityDate, _ := time.Parse("2006-01-02", res.Security.MaturityDate)

	return &securitypb.Security{
		Id:           res.Security.ID,
		Isin:         res.Security.Isin,
		Name:         res.Security.Name,
		SecurityType: res.Security.Type,
		Serie:        res.Security.Serie,
		IssueDate:    timestamppb.New(issueDate),
		MaturityDate: timestamppb.New(maturityDate),
		NominalValue: parseFloat(res.Security.NominalValue),
		Currency:     res.Security.Currency,
		Fx:           parseFloat(res.Security.Fx),
		Enable:       res.Security.Enable,
	}, nil
}

func (h *SecurityGRPCHandler) GetEnableSecurities(ctx context.Context, _ *securitypb.Empty) (*securitypb.ListSecuritiesResponse, error) {
	res, err := h.getEnable.Execute(ctx)
	if err != nil {
		return nil, err
	}

	securities := make([]*securitypb.Security, 0, len(res.Results))
	for _, s := range res.Results {
		issueDate, _ := time.Parse("2006-01-02", s.IssueDate)
		maturityDate, _ := time.Parse("2006-01-02", s.MaturityDate)
		securities = append(securities, &securitypb.Security{
			Id:           s.ID,
			Isin:         s.Isin,
			Name:         s.Name,
			SecurityType: s.Type,
			Serie:        s.Serie,
			IssueDate:    timestamppb.New(issueDate),
			MaturityDate: timestamppb.New(maturityDate),
			NominalValue: parseFloat(s.NominalValue),
			Currency:     s.Currency,
			Fx:           parseFloat(s.Fx),
			Enable:       s.Enable,
		})
	}

	return &securitypb.ListSecuritiesResponse{Securities: securities}, nil
}


func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
