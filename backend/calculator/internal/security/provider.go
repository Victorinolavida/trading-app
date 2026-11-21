package security

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/victorinolavida/calculator/api"
	securitypb "github.com/victorinolavida/proto/security"
	"strconv"
)

type Provider struct{ Client securitypb.AuthServiceClient }

func (p Provider) GetByPublicID(ctx context.Context, id uuid.UUID) (api.Security, error) {
	s, err := p.Client.GetByParams(ctx, &securitypb.ListSecuritiesRequest{Id: id.String()})
	if err != nil {
		return api.Security{}, err
	}
	if s.MaturityDate == nil || !s.MaturityDate.IsValid() {
		return api.Security{}, fmt.Errorf("security maturity date is invalid")
	}
	nominal, err := api.ParseDecimal(strconv.FormatFloat(s.NominalValue, 'f', -1, 64))
	if err != nil {
		return api.Security{}, err
	}
	return api.NewBondSecurity(s.Id, s.Isin, s.Name, s.SecurityType, s.Currency, s.MaturityDate.AsTime(), nominal), nil
}
