// Package api is the public boundary of the orders bounded context.
package api

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/victorinolavida/orders/internal/adapters/out/postgres"
	orderproducer "github.com/victorinolavida/orders/internal/adapters/out/producer/rabbitmq"
	redisadapter "github.com/victorinolavida/orders/internal/adapters/out/redis"
	"github.com/victorinolavida/orders/internal/application/ports"
	matchingengine "github.com/victorinolavida/orders/internal/application/services/matching-engine"
	"github.com/victorinolavida/orders/internal/application/usecases/active_orders"
	"github.com/victorinolavida/orders/internal/application/usecases/cancel_order"
	"github.com/victorinolavida/orders/internal/application/usecases/create_order"
	"github.com/victorinolavida/orders/internal/application/usecases/get_market_depth"
	"github.com/victorinolavida/orders/internal/application/usecases/list_orders"
	orderexpiration "github.com/victorinolavida/orders/internal/application/workers/order_expiration"
	"github.com/victorinolavida/orders/internal/domain"
	broker "github.com/victorinolavida/pkg/broker/rabbitmq"
	pkgconfig "github.com/victorinolavida/pkg/config"
)

// Order is the order representation exposed across module boundaries.
type Order struct {
	ID                string
	SecurityID        string
	Direction         string
	OrderType         string
	Price             string
	Yield             string
	Quantity          int64
	NominalValue      string
	AllocatedQuantity int64
	AllocationPtc     string
	AllocationStatus  string
	Fx                string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ExpiresAt         time.Time
}

// MarketDepthLevel is one aggregated price level of the order book.
type MarketDepthLevel struct {
	Price        string
	NominalValue string
	Quantity     int64
	OrderCounts  int64
	Direction    string
}

// MarketDepth is the aggregated order book of a security.
type MarketDepth struct {
	Bids []MarketDepthLevel
	Asks []MarketDepthLevel
}

// UserInfo is the user data the orders module needs from the users context.
type UserInfo struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

// UserProvider is the dependency the orders module needs from the users
// context. It is declared here so orders does not import the users module.
type UserProvider interface {
	GetByID(ctx context.Context, id uuid.UUID) (UserInfo, error)
	Whoami(ctx context.Context, id uuid.UUID) (UserInfo, error)
}

// SecurityInfo is the security data the orders module needs from the
// securities context.
type SecurityInfo struct {
	ID           uuid.UUID
	ISIN         string
	Name         string
	SecurityType string
	Serie        string
	NominalValue float64
	Currency     string
	Fx           float64
	Enable       bool
}

// SecurityProvider is the dependency the orders module needs from the
// securities context. It is declared here so orders does not import the
// securities module.
type SecurityProvider interface {
	GetByPublicID(ctx context.Context, id uuid.UUID) (SecurityInfo, error)
	GetByISIN(ctx context.Context, isin string) (SecurityInfo, error)
}

// CreateOrderCommand is the input of the create order use case.
type CreateOrderCommand struct {
	TraderID   string
	SecurityID string
	Direction  string
	OrderType  string
	Price      *string
	Yield      *string
	Quantity   int64
	Expiration time.Duration
}

// CancelOrderCommand is the input of the cancel order use case.
type CancelOrderCommand struct {
	OrderID  string
	TraderID string
}

// ListOrdersQuery filters the trader order listing.
type ListOrdersQuery struct {
	TraderID  string
	OrderIDs  []string
	Direction string
	Status    string
}

// ActiveOrdersQuery filters the active orders of a security.
type ActiveOrdersQuery struct {
	TraderID   string
	SecurityID string
	Direction  string
}

// Deps are the external dependencies the orders module needs to start.
type Deps struct {
	Pool                  *pgxpool.Pool
	Redis                 pkgconfig.RedisConfig
	RabbitMQ              pkgconfig.RabbitMQ
	Exchange              string
	AllowInternalCrossing bool
	Users                 UserProvider
	Securities            SecurityProvider
}

// Module is the in-process orders module.
type Module struct {
	rabbit         *broker.Client
	createOrder    *create_order.Usecase
	cancelOrder    *cancel_order.Usecase
	activeOrders   *active_orders.Usecase
	listOrders     *list_orders.Usecase
	getMarketDepth *get_market_depth.Usecase
	worker         *orderexpiration.OrderWorker
}

// New wires the orders module and its infrastructure.
func New(deps Deps) (*Module, error) {
	memoryPort, err := redisadapter.New(context.Background(), deps.Redis)
	if err != nil {
		return nil, err
	}

	rabbitClient, err := broker.NewClientFromConfig(deps.RabbitMQ)
	if err != nil {
		return nil, err
	}

	publisher, err := orderproducer.NewProducer(memoryPort, rabbitClient, deps.Exchange)
	if err != nil {
		rabbitClient.Close()
		return nil, err
	}

	orderRepo := postgres.NewOrderRepository(deps.Pool)
	tradeRepo := postgres.NewTradeRepository(deps.Pool)
	txManager := postgres.NewTxManager(deps.Pool)

	var crossingPolicy domain.InternalCrossingPolicy = domain.SameTraderCrossingPolicy{Action: domain.CrossingCancelTaker}
	if deps.AllowInternalCrossing {
		crossingPolicy = domain.AllowCrossingPolicy{}
	}

	engine := matchingengine.New(orderRepo, crossingPolicy)

	return &Module{
		rabbit:         rabbitClient,
		createOrder:    create_order.New(orderRepo, tradeRepo, securityAdapter{deps.Securities}, userAdapter{deps.Users}, engine, txManager, publisher),
		cancelOrder:    cancel_order.New(orderRepo, publisher, txManager),
		activeOrders:   active_orders.New(orderRepo),
		listOrders:     list_orders.New(orderRepo),
		getMarketDepth: get_market_depth.New(orderRepo),
		worker:         orderexpiration.NewOrderWorker(orderRepo, txManager, publisher, 10*time.Second),
	}, nil
}

// CreateOrder creates a new order and matches it against the book.
func (m *Module) CreateOrder(ctx context.Context, command CreateOrderCommand) (Order, error) {
	traderID, err := uuid.Parse(command.TraderID)
	if err != nil {
		return Order{}, err
	}
	securityID, err := uuid.Parse(command.SecurityID)
	if err != nil {
		return Order{}, err
	}

	price, err := parseScaled(command.Price)
	if err != nil {
		return Order{}, err
	}
	yield, err := parseScaled(command.Yield)
	if err != nil {
		return Order{}, err
	}

	res, err := m.createOrder.Execute(ctx, create_order.Command{
		TraderID:   traderID,
		SecurityID: securityID,
		Direction:  command.Direction,
		OrderType:  command.OrderType,
		Price:      price,
		Yield:      yield,
		Quantity:   command.Quantity,
		ExpiresAt:  command.Expiration,
	})
	if err != nil {
		return Order{}, err
	}
	return fromCreateResult(res), nil
}

// CancelOrder cancels an active order owned by the trader.
func (m *Module) CancelOrder(ctx context.Context, command CancelOrderCommand) (Order, error) {
	orderID, err := uuid.Parse(command.OrderID)
	if err != nil {
		return Order{}, err
	}
	traderID, err := uuid.Parse(command.TraderID)
	if err != nil {
		return Order{}, err
	}

	res, err := m.cancelOrder.Execute(ctx, cancel_order.Command{OrderID: orderID, TraderID: traderID})
	if err != nil {
		return Order{}, err
	}
	return fromCancelResult(res), nil
}

// ListOrders returns the orders of a trader.
func (m *Module) ListOrders(ctx context.Context, query ListOrdersQuery) ([]Order, error) {
	traderID, err := uuid.Parse(query.TraderID)
	if err != nil {
		return nil, err
	}

	orderIDs := make([]uuid.UUID, 0, len(query.OrderIDs))
	for _, raw := range query.OrderIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, err
		}
		orderIDs = append(orderIDs, id)
	}

	res, err := m.listOrders.Execute(ctx, list_orders.Command{
		TraderID:  traderID,
		OrderIDs:  orderIDs,
		Direction: query.Direction,
		Status:    query.Status,
	})
	if err != nil {
		return nil, err
	}

	out := make([]Order, len(res.Orders))
	for i, o := range res.Orders {
		out[i] = Order{
			ID:                o.ID,
			SecurityID:        o.SecurityId,
			Direction:         o.Direction,
			OrderType:         o.OrderType,
			Price:             o.Price,
			Yield:             o.Yield,
			Quantity:          o.Quantity,
			NominalValue:      o.NominalValue,
			AllocatedQuantity: o.AllocatedQuantity,
			AllocationPtc:     o.AllocationPtc,
			AllocationStatus:  o.AllocationStatus,
			Fx:                o.Fx,
			Status:            o.Status,
			CreatedAt:         o.CreatedAt,
			UpdatedAt:         o.UpdatedAt,
			ExpiresAt:         o.ExpiresAt,
		}
	}
	return out, nil
}

// ActiveOrders returns the active orders of a security for a trader.
func (m *Module) ActiveOrders(ctx context.Context, query ActiveOrdersQuery) ([]Order, error) {
	traderID, err := uuid.Parse(query.TraderID)
	if err != nil {
		return nil, err
	}
	securityID, err := uuid.Parse(query.SecurityID)
	if err != nil {
		return nil, err
	}

	res, err := m.activeOrders.Execute(ctx, active_orders.Command{
		TraderID:   traderID,
		SecurityId: securityID,
		Direction:  query.Direction,
	})
	if err != nil {
		return nil, err
	}

	out := make([]Order, len(res.Orders))
	for i, o := range res.Orders {
		out[i] = Order{
			ID:                o.ID,
			SecurityID:        o.SecurityId,
			Direction:         o.Direction,
			OrderType:         o.OrderType,
			Price:             o.Price,
			Yield:             o.Yield,
			Quantity:          o.Quantity,
			NominalValue:      o.NominalValue,
			AllocatedQuantity: o.AllocatedQuantity,
			AllocationPtc:     o.AllocationPtc,
			AllocationStatus:  o.AllocationStatus,
			Fx:                o.Fx,
			Status:            o.Status,
			CreatedAt:         o.CreatedAt,
			UpdatedAt:         o.UpdatedAt,
			ExpiresAt:         o.ExpiresAt,
		}
	}
	return out, nil
}

// MarketDepth returns the aggregated order book of a security.
func (m *Module) MarketDepth(ctx context.Context, securityID uuid.UUID) (*MarketDepth, error) {
	res, err := m.getMarketDepth.Execute(ctx, get_market_depth.Command{SecurityID: securityID})
	if err != nil {
		return nil, err
	}
	if res.MarketDepth == nil {
		return &MarketDepth{}, nil
	}
	return &MarketDepth{
		Bids: toDepthLevels(res.MarketDepth.Bid),
		Asks: toDepthLevels(res.MarketDepth.Ask),
	}, nil
}

// StartWorker runs the order expiration worker until the context is canceled.
func (m *Module) StartWorker(ctx context.Context) error {
	return m.worker.Start(ctx)
}

// Close releases the orders module infrastructure.
func (m *Module) Close() error {
	return m.rabbit.Close()
}

func toDepthLevels(sides []*domain.MarketDepthSide) []MarketDepthLevel {
	out := make([]MarketDepthLevel, 0, len(sides))
	for _, side := range sides {
		if side == nil {
			continue
		}
		out = append(out, MarketDepthLevel{
			Price:        side.Price.String(),
			NominalValue: side.AggNominalValue.String(),
			Quantity:     side.AggQuantity.Value(),
			OrderCounts:  side.OrderCounts,
			Direction:    side.Direction.String(),
		})
	}
	return out
}

func parseScaled(value *string) (int64, error) {
	if value == nil || *value == "" {
		return 0, nil
	}
	d, err := decimal.NewFromString(*value)
	if err != nil {
		return 0, err
	}
	return d.Mul(decimal.NewFromInt(int64(domain.DecimalScale))).IntPart(), nil
}

func toDecimal(value float64) domain.Decimal {
	return domain.Decimal(math.Round(value * float64(domain.DecimalScale)))
}

func fromCreateResult(res create_order.Result) Order {
	return Order{
		ID:                res.ID,
		SecurityID:        res.SecurityId,
		Direction:         res.Direction,
		OrderType:         res.OrderType,
		Price:             res.Price,
		Yield:             res.Yield,
		Quantity:          res.Quantity,
		NominalValue:      res.NominalValue,
		AllocatedQuantity: res.AllocatedQuantity,
		AllocationPtc:     res.AllocationPtc,
		AllocationStatus:  res.AllocationStatus,
		Fx:                res.Fx,
		Status:            res.Status,
		CreatedAt:         res.CreatedAt,
		UpdatedAt:         res.UpdatedAt,
		ExpiresAt:         res.ExpiresAt,
	}
}

func fromCancelResult(res cancel_order.Result) Order {
	return Order{
		ID:                res.ID,
		SecurityID:        res.SecurityId,
		Direction:         res.Direction,
		OrderType:         res.OrderType,
		Price:             res.Price,
		Yield:             res.Yield,
		Quantity:          res.Quantity,
		NominalValue:      res.NominalValue,
		AllocatedQuantity: res.AllocatedQuantity,
		AllocationPtc:     res.AllocationPtc,
		AllocationStatus:  res.AllocationStatus,
		Fx:                res.Fx,
		Status:            res.Status,
		CreatedAt:         res.CreatedAt,
		UpdatedAt:         res.UpdatedAt,
		ExpiresAt:         res.ExpiresAt,
	}
}

type userAdapter struct {
	provider UserProvider
}

var _ ports.UserService = userAdapter{}

func (a userAdapter) GetByID(ctx context.Context, id uuid.UUID) (*ports.UserInfo, error) {
	user, err := a.provider.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPortUser(user), nil
}

func (a userAdapter) Whoami(ctx context.Context, id uuid.UUID) (*ports.UserInfo, error) {
	user, err := a.provider.Whoami(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPortUser(user), nil
}

func toPortUser(user UserInfo) *ports.UserInfo {
	return &ports.UserInfo{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
}

type securityAdapter struct {
	provider SecurityProvider
}

var _ ports.SecurityService = securityAdapter{}

func (a securityAdapter) GetByID(ctx context.Context, id uuid.UUID) (*ports.SecurityInfo, error) {
	security, err := a.provider.GetByPublicID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toPortSecurity(id, security), nil
}

func (a securityAdapter) GetByISIN(ctx context.Context, isin string) (*ports.SecurityInfo, error) {
	security, err := a.provider.GetByISIN(ctx, isin)
	if err != nil {
		return nil, err
	}
	return toPortSecurity(uuid.Nil, security), nil
}

func toPortSecurity(id uuid.UUID, security SecurityInfo) *ports.SecurityInfo {
	return &ports.SecurityInfo{
		ID:           id,
		PublicID:     id,
		ISIN:         security.ISIN,
		Name:         security.Name,
		SecurityType: security.SecurityType,
		Serie:        security.Serie,
		NominalValue: toDecimal(security.NominalValue),
		Currency:     security.Currency,
		Fx:           toDecimal(security.Fx),
		Enable:       security.Enable,
	}
}
