package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domain"
)

type SecurityInfo struct {
	ID           uuid.UUID
	PublicID     uuid.UUID
	ISIN         string
	Name         string
	SecurityType string
	Serie        string
	NominalValue domain.Decimal
	Currency     string
	Fx           domain.Decimal
	Enable       bool
}

type SecurityService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*SecurityInfo, error)
	GetByISIN(ctx context.Context, isin string) (*SecurityInfo, error)
}

type UserInfo struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

type UserService interface {
	GetByID(ctx context.Context, id uuid.UUID) (*UserInfo, error)
	Whoami(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
}
