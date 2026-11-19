package main

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	usersapi "github.com/victorinolavida/auth/api"
	ordersapi "github.com/victorinolavida/orders/api"
	securitiesapi "github.com/victorinolavida/securities/api"
)

// userProvider adapts the public users module to the dependency the orders
// module declares.
type userProvider struct {
	module *usersapi.Module
}

var _ ordersapi.UserProvider = userProvider{}

func (p userProvider) GetByID(ctx context.Context, id uuid.UUID) (ordersapi.UserInfo, error) {
	user, err := p.module.GetByID(ctx, id)
	if err != nil {
		return ordersapi.UserInfo{}, err
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return ordersapi.UserInfo{}, err
	}
	return ordersapi.UserInfo{
		ID:        userID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (p userProvider) Whoami(ctx context.Context, id uuid.UUID) (ordersapi.UserInfo, error) {
	return p.GetByID(ctx, id)
}

// securityProvider adapts the public securities module to the dependency the
// orders module declares.
type securityProvider struct {
	module *securitiesapi.Module
}

var _ ordersapi.SecurityProvider = securityProvider{}

func (p securityProvider) GetByPublicID(ctx context.Context, id uuid.UUID) (ordersapi.SecurityInfo, error) {
	security, err := p.module.GetByPublicID(ctx, id)
	if err != nil {
		return ordersapi.SecurityInfo{}, err
	}
	return toOrdersSecurity(id, security), nil
}

func (p securityProvider) GetByISIN(ctx context.Context, isin string) (ordersapi.SecurityInfo, error) {
	security, err := p.module.GetByISIN(ctx, isin)
	if err != nil {
		return ordersapi.SecurityInfo{}, err
	}
	id, _ := uuid.Parse(security.ID)
	return toOrdersSecurity(id, security), nil
}

func toOrdersSecurity(id uuid.UUID, security *securitiesapi.Security) ordersapi.SecurityInfo {
	nominalValue, _ := strconv.ParseFloat(security.NominalValue, 64)
	fx, _ := strconv.ParseFloat(security.Fx, 64)
	return ordersapi.SecurityInfo{
		ID:           id,
		ISIN:         security.ISIN,
		Name:         security.Name,
		SecurityType: security.SecurityType,
		Serie:        security.Serie,
		NominalValue: nominalValue,
		Currency:     security.Currency,
		Fx:           fx,
		Enable:       security.Enable,
	}
}
