package securities

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
	securitypb "github.com/victorinolavida/proto/security"
	"google.golang.org/grpc"
)

func toDecimal(v float64) domain.Decimal {
	return domain.Decimal(math.Round(v * float64(domain.DecimalScale)))
}

type Service struct {
	client securitypb.AuthServiceClient
}

func New(conn *grpc.ClientConn) *Service {
	return &Service{client: securitypb.NewAuthServiceClient(conn)}
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*ports.SecurityInfo, error) {
	res, err := s.client.GetByParams(ctx, &securitypb.ListSecuritiesRequest{Id: id.String()})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("security not found")
	}
	secID, err := uuid.Parse(res.Id)
	if err != nil {
		secID = id
	}
	pubID := secID
	return &ports.SecurityInfo{
		ID:           secID,
		PublicID:     pubID,
		ISIN:         res.Isin,
		Name:         res.Name,
		SecurityType: res.SecurityType,
		Serie:        res.Serie,
		NominalValue: toDecimal(res.NominalValue),
		Currency:     res.Currency,
		Fx:           toDecimal(res.Fx),
		Enable:       res.Enable,
	}, nil
}

func (s *Service) GetByISIN(ctx context.Context, isin string) (*ports.SecurityInfo, error) {
	res, err := s.client.GetByParams(ctx, &securitypb.ListSecuritiesRequest{Isin: isin})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("security not found")
	}
	secID, err := uuid.Parse(res.Id)
	if err != nil {
		return nil, err
	}
	pubID := secID
	return &ports.SecurityInfo{
		ID:           secID,
		PublicID:     pubID,
		ISIN:         res.Isin,
		Name:         res.Name,
		SecurityType: res.SecurityType,
		Serie:        res.Serie,
		NominalValue: toDecimal(res.NominalValue),
		Currency:     res.Currency,
		Fx:           toDecimal(res.Fx),
		Enable:       res.Enable,
	}, nil
}
