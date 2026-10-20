package ports

import (
	"context"
	"users/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(context.Context, domain.User) error
	GetByID(context.Context, uuid.UUID) (domain.User, error)
	GetByEmail(context.Context, domain.Email) (domain.User, error)
	ExistByEmail(context.Context, domain.Email) (bool, error)
	Update(context.Context, domain.User) error
}
