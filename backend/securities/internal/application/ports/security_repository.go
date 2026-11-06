package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/securities/internal/domain"
)

type TimeFilters struct {
	IssueBefore    time.Time
	IssueAfter     time.Time
	MaturityAfter  time.Time
	MaturityBefore time.Time
}

type Params struct {
	IDs           uuid.UUIDs
	ISINs         []string
	Serie         []string
	SecurityTypes []domain.SecurityType
	TimeFilters   TimeFilters
	Name          string
	Query         string
	ISIN          string
}

type SecurityRepository interface {
	GetByID(context.Context, uuid.UUID) (*domain.Security, error)
	GetByISIN(context.Context, string) (*domain.Security, error)
	GetListByParams(context.Context, Params) ([]*domain.Security, error)
	GetEnableList(context.Context) ([]*domain.Security, error)
}
