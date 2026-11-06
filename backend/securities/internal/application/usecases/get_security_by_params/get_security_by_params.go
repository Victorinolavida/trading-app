package getsecuritybyparams

import (
	"context"

	"github.com/google/uuid"
	"github.com/victorinolavida/securities/internal/application/ports"
	"github.com/victorinolavida/securities/internal/domain"
)

type Usecase struct {
	repo ports.SecurityRepository
}

func New(repo ports.SecurityRepository) *Usecase {
	return &Usecase{repo}
}

func (u *Usecase) Execute(ctx context.Context, cmd Command) (Result, error) {
	var sec *domain.Security
	var err error

	if cmd.ID != "" {
		id, err := uuid.Parse(cmd.ID)
		if err != nil {
			return Result{}, err
		}
		sec, err = u.repo.GetByID(ctx, id)
		if err != nil {
			return Result{}, err
		}
	} else if cmd.ISIN != "" {
		sec, err = u.repo.GetByISIN(ctx, cmd.ISIN)
		if err != nil {
			return Result{}, err
		}
	} else {
		filters := ports.Params{
			TimeFilters: ports.TimeFilters{
				IssueBefore:    cmd.IssueBefore,
				IssueAfter:     cmd.IssueAfter,
				MaturityBefore: cmd.MaturityBefore,
				MaturityAfter:  cmd.MaturityAfter,
			},
		}
		if cmd.Serie != "" {
			filters.Serie = []string{cmd.Serie}
		}
		if cmd.SecurityType != "" {
			filters.SecurityTypes = []domain.SecurityType{domain.SecurityType(cmd.SecurityType)}
		}
		securities, err := u.repo.GetListByParams(ctx, filters)
		if err != nil {
			return Result{}, err
		}
		if len(securities) == 0 {
			return Result{}, nil
		}
		sec = securities[0]
	}

	return Result{
		Security: Security{
			ID:           sec.PublicID().String(),
			Isin:         sec.ISIN(),
			Name:         sec.Name(),
			Serie:        sec.Serie(),
			Type:         sec.SecurityType().String(),
			IssueDate:    sec.IssueDate().Format("2006-01-02"),
			MaturityDate: sec.MaturityDate().Format("2006-01-02"),
			NominalValue: sec.NominalValue().String(),
			Currency:     sec.Currency().String(),
			Fx:           sec.Fx().String(),
			Enable:       sec.Enable(),
		},
	}, nil
}
