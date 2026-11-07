package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/orders/internal/domain"
)

type MockOrderRepo struct {
	mock.Mock
}

var _ ports.OrderRepository = (*MockOrderRepo)(nil)

func NewMockOrderRepo() *MockOrderRepo {
	return &MockOrderRepo{}
}

func (m *MockOrderRepo) Save(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepo) GetByPublicID(ctx context.Context, publicID uuid.UUID) (*domain.Order, error) {
	args := m.Called(ctx, publicID)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepo) Update(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) UpdateBulk(ctx context.Context, orders ...*domain.Order) error {
	args := m.Called(ctx, orders)
	return args.Error(0)
}

func (m *MockOrderRepo) GetOrdersToTrade(ctx context.Context, order *domain.Order) ([]*domain.Order, error) {
	args := m.Called(ctx, order)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepo) GetActiveOrders(ctx context.Context, traderID uuid.UUID, securityID uuid.UUID, direction string) ([]*domain.Order, error) {
	args := m.Called(ctx, traderID, securityID, direction)
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepo) ListOrders(ctx context.Context, filter ports.ListOrdersFilter) ([]*domain.Order, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.Order), args.Error(1)
}
