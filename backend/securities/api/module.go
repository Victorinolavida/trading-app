// Package api is the public boundary of the securities bounded context.
package api

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/securities/internal/adapters/out/postgres"
	"github.com/victorinolavida/securities/internal/application/ports"
	getenablesecurities "github.com/victorinolavida/securities/internal/application/usecases/get_enable_securities"
	getsecurities "github.com/victorinolavida/securities/internal/application/usecases/get_securities"
	getsecuritybyparams "github.com/victorinolavida/securities/internal/application/usecases/get_security_by_params"
	"github.com/victorinolavida/securities/internal/domain"
	"github.com/victorinolavida/securities/internal/domainerrors"
)

// ErrSecurityNotFound is re-exported for callers that need to classify the
// failure without importing the internal domain error package.
var ErrSecurityNotFound = domainerrors.ErrSecurityNotFound

// Security is the security representation exposed across module boundaries.
type Security struct {
	ID           string
	ISIN         string
	Name         string
	SecurityType string
	Serie        string
	IssueDate    string
	MaturityDate string
	NominalValue string
	Currency     string
	Fx           string
	Enable       bool
}

// Params mirrors the list filters of the securities module.
type Params struct {
	IDs            []uuid.UUID
	ISINs          []string
	Serie          []string
	SecurityTypes  []string
	Name           string
	Query          string
	ISIN           string
	IssueBefore    time.Time
	IssueAfter     time.Time
	MaturityBefore time.Time
	MaturityAfter  time.Time
}

// GetParams mirrors the lookup filters of a single security.
type GetParams struct {
	ID             string
	ISIN           string
	Serie          string
	SecurityType   string
	IssueBefore    time.Time
	IssueAfter     time.Time
	MaturityBefore time.Time
	MaturityAfter  time.Time
}

// Module is the in-process securities module.
type Module struct {
	getList     *getsecurities.Usecase
	getEnabled  *getenablesecurities.Usecase
	getByParams *getsecuritybyparams.Usecase
}

// New wires the securities module on top of the given database pool.
func New(pool *pgxpool.Pool) *Module {
	repo := postgres.NewSecurityRepo(pool)
	return &Module{
		getList:     getsecurities.New(repo),
		getEnabled:  getenablesecurities.New(repo),
		getByParams: getsecuritybyparams.New(repo),
	}
}

// List returns the securities matching the given filters.
func (m *Module) List(ctx context.Context, params Params) ([]Security, error) {
	res, err := m.getList.Execute(ctx, toPortParams(params))
	if err != nil {
		return nil, err
	}
	return toSecurities(res.Results), nil
}

// Get returns a single security matching the given filters.
func (m *Module) Get(ctx context.Context, params GetParams) (*Security, error) {
	if params.ID != "" {
		id, err := uuid.Parse(params.ID)
		if err != nil {
			return nil, err
		}
		return m.GetByPublicID(ctx, id)
	}

	res, err := m.getByParams.Execute(ctx, getsecuritybyparams.Command{
		ISIN:           params.ISIN,
		Serie:          params.Serie,
		SecurityType:   params.SecurityType,
		IssueBefore:    params.IssueBefore,
		IssueAfter:     params.IssueAfter,
		MaturityBefore: params.MaturityBefore,
		MaturityAfter:  params.MaturityAfter,
	})
	if err != nil {
		return nil, err
	}
	if res.Security.ID == "" && res.Security.Isin == "" {
		return nil, ErrSecurityNotFound
	}
	sec := fromGetSecurity(res.Security)
	return &sec, nil
}

// GetEnabled returns every enabled security.
func (m *Module) GetEnabled(ctx context.Context) ([]Security, error) {
	res, err := m.getEnabled.Execute(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Security, len(res.Results))
	for i, s := range res.Results {
		out[i] = fromEnabledSecurity(s)
	}
	return out, nil
}

// GetByPublicID returns a security by its public id.
func (m *Module) GetByPublicID(ctx context.Context, id uuid.UUID) (*Security, error) {
	res, err := m.getList.Execute(ctx, ports.Params{IDs: uuid.UUIDs{id}})
	if err != nil {
		return nil, err
	}
	if len(res.Results) == 0 {
		return nil, ErrSecurityNotFound
	}
	sec := fromListSecurity(res.Results[0])
	return &sec, nil
}

// GetByISIN returns a security by its ISIN.
func (m *Module) GetByISIN(ctx context.Context, isin string) (*Security, error) {
	res, err := m.getByParams.Execute(ctx, getsecuritybyparams.Command{ISIN: isin})
	if err != nil {
		return nil, err
	}
	if res.Security.ID == "" && res.Security.Isin == "" {
		return nil, ErrSecurityNotFound
	}
	sec := fromGetSecurity(res.Security)
	return &sec, nil
}

func toPortParams(params Params) ports.Params {
	types := make([]domain.SecurityType, len(params.SecurityTypes))
	for i, t := range params.SecurityTypes {
		types[i] = domain.SecurityType(t)
	}
	return ports.Params{
		IDs:           params.IDs,
		ISINs:         params.ISINs,
		Serie:         params.Serie,
		SecurityTypes: types,
		Name:          params.Name,
		Query:         params.Query,
		ISIN:          params.ISIN,
		TimeFilters: ports.TimeFilters{
			IssueBefore:    params.IssueBefore,
			IssueAfter:     params.IssueAfter,
			MaturityBefore: params.MaturityBefore,
			MaturityAfter:  params.MaturityAfter,
		},
	}
}

func toSecurities(list []getsecurities.Security) []Security {
	out := make([]Security, len(list))
	for i, s := range list {
		out[i] = fromListSecurity(s)
	}
	return out
}

func fromListSecurity(s getsecurities.Security) Security {
	return Security{
		ID:           s.ID,
		ISIN:         s.Isin,
		Name:         s.Name,
		SecurityType: s.Type,
		Serie:        s.Serie,
		IssueDate:    s.IssueDate,
		MaturityDate: s.MaturityDate,
		NominalValue: s.NominalValue,
		Currency:     s.Currency,
		Fx:           s.Fx,
		Enable:       s.Enable,
	}
}

func fromGetSecurity(s getsecuritybyparams.Security) Security {
	return Security{
		ID:           s.ID,
		ISIN:         s.Isin,
		Name:         s.Name,
		SecurityType: s.Type,
		Serie:        s.Serie,
		IssueDate:    s.IssueDate,
		MaturityDate: s.MaturityDate,
		NominalValue: s.NominalValue,
		Currency:     s.Currency,
		Fx:           s.Fx,
		Enable:       s.Enable,
	}
}

// fromEnabledSecurity maps the enable-list use case result, which uses the same
// field layout as the list use case.
func fromEnabledSecurity(s getenablesecurities.Security) Security {
	return Security{
		ID:           s.ID,
		ISIN:         s.Isin,
		Name:         s.Name,
		SecurityType: s.Type,
		Serie:        s.Serie,
		IssueDate:    s.IssueDate,
		MaturityDate: s.MaturityDate,
		NominalValue: s.NominalValue,
		Currency:     s.Currency,
		Fx:           s.Fx,
		Enable:       s.Enable,
	}
}
