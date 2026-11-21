package create_order

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/domain"
	"github.com/victorinolavida/orders/internal/domainerrors"
)

type Calculator interface {
	Calculate(ctx context.Context, securityID, settlement, field, value string) (string, error)
}

func (u *Usecase) completeLimitOrder(ctx context.Context, cmd *Command) error {
	if cmd.OrderType != string(domain.OrderTypeLimit) {
		return nil
	}
	if cmd.Price < 0 || (cmd.Price == 0 && cmd.Yield == 0) {
		return domainerrors.ErrInvalidPrice
	}
	if cmd.Yield < 0 {
		return domainerrors.ErrInvalidYield
	}
	if cmd.Price != 0 && cmd.Yield != 0 {
		return nil
	}
	if u.calculator == nil {
		return errors.New("calculator unavailable")
	}
	now := time.Now().UTC()
	settlement := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	field, value := "yield", domain.Decimal(cmd.Yield).String()
	if cmd.Price != 0 {
		field, value = "clean_price", domain.Decimal(cmd.Price).String()
	}
	result, err := u.calculator.Calculate(ctx, cmd.SecurityID.String(), settlement, field, value)
	if err != nil {
		return err
	}
	n, err := decimal.NewFromString(result)
	if err != nil || n.IsNegative() {
		return domain.ErrInvalidNumericNumber
	}
	scaled := n.Mul(decimal.NewFromInt(domain.DecimalScale)).Round(0)
	if scaled.GreaterThan(decimal.NewFromInt(domain.MaxDecimalValue)) {
		return domain.ErrInvalidNumericNumber
	}
	if field == "yield" {
		cmd.Price = scaled.IntPart()
		if cmd.Price <= 0 {
			return domainerrors.ErrInvalidPrice
		}
	} else {
		cmd.Yield = scaled.IntPart()
	}
	return nil
}
