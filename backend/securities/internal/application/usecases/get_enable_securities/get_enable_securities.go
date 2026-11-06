package getenablesecurities

import (
	"context"

	"github.com/victorinolavida/securities/internal/application/ports"
)

type Usecase struct {
	repo ports.SecurityRepository
}

func New(repo ports.SecurityRepository) *Usecase {
	return &Usecase{repo}
}

func (u *Usecase) Execute(ctx context.Context) (Result, error) {
	securities, err := u.repo.GetEnableList(ctx)
	if err != nil {
		return Result{}, nil
	}

	result := Result{}
	sec := make([]Security, 0, len(securities))
	for _, security := range securities {
		sec = append(sec, Security{
			ID:           security.PublicID().String(),
			Isin:         security.ISIN(),
			Name:         security.Name(),
			Serie:        security.Serie(),
			Type:         security.SecurityType().String(),
			IssueDate:    security.IssueDate().Format("2006-01-02"),
			MaturityDate: security.MaturityDate().Format("2006-01-02"),
			NominalValue: security.NominalValue().String(),
			Currency:     security.Currency().String(),
			Fx:           security.Fx().String(),
			Enable:       security.Enable(),
		})
	}

	result.Results = sec
	return result, nil
}
