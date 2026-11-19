package api

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	usersapi "github.com/victorinolavida/auth/api"
	calculatorapi "github.com/victorinolavida/calculator/api"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	ordersapi "github.com/victorinolavida/orders/api"
	securitiesapi "github.com/victorinolavida/securities/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// authAdapter implements the gateway auth port on top of the in-process users
// module. It translates module errors to the same gRPC codes the gateway used
// to receive from the users service, so the HTTP error mapping stays intact.
type authAdapter struct {
	module *usersapi.Module
}

var _ ports.AuthService = authAdapter{}

func (a authAdapter) Login(ctx context.Context, req dtos.LoginRequest) (dtos.AuthUserResponse, error) {
	res, err := a.module.Login(ctx, req.Email, req.Password)
	if err != nil {
		return dtos.AuthUserResponse{}, mapUserError(err)
	}
	return dtos.AuthUserResponse{
		UserID:    res.User.ID,
		FirstName: res.User.FirstName,
		LastName:  res.User.LastName,
		Token:     res.Token,
	}, nil
}

func (a authAdapter) Register(ctx context.Context, req dtos.RegisterRequest) (dtos.AuthUserResponse, error) {
	res, err := a.module.Register(ctx, req.Email, req.FirstName, req.LastName, req.Password)
	if err != nil {
		return dtos.AuthUserResponse{}, mapUserError(err)
	}
	return dtos.AuthUserResponse{
		UserID:    res.User.ID,
		FirstName: res.User.FirstName,
		LastName:  res.User.LastName,
		Token:     res.Token,
	}, nil
}

func (a authAdapter) Whoami(ctx context.Context, req dtos.WhoamiUserRequest) (dtos.WhoamiUserResponse, error) {
	user, err := a.module.Whoami(ctx, req.UserID)
	if err != nil {
		return dtos.WhoamiUserResponse{}, mapUserError(err)
	}
	return dtos.WhoamiUserResponse{
		UserID:    user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func mapUserError(err error) error {
	switch {
	case errors.Is(err, usersapi.ErrInvalidEmail),
		errors.Is(err, usersapi.ErrNameTooShort),
		errors.Is(err, usersapi.ErrNameTooLong),
		errors.Is(err, usersapi.ErrPasswordTooShort),
		errors.Is(err, usersapi.ErrPasswordTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usersapi.ErrEmailAlreadyRegister):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, usersapi.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case errors.Is(err, usersapi.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	case errors.Is(err, usersapi.ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

// securitiesProviderAdapter implements the calculator's SecurityProvider interface
// on top of the in-process securities module.
type securitiesProviderAdapter struct {
	module *securitiesapi.Module
}

var _ calculatorapi.SecurityProvider = securitiesProviderAdapter{}

func (a securitiesProviderAdapter) GetByPublicID(ctx context.Context, publicID uuid.UUID) (calculatorapi.Security, error) {
	sec, err := a.module.GetByPublicID(ctx, publicID)
	if err != nil {
		return calculatorapi.Security{}, err
	}
	maturityDate, err := time.Parse("2006-01-02", sec.MaturityDate)
	if err != nil {
		maturityDate = time.Time{}
	}
	return calculatorapi.Security{
		ID:               sec.ID,
		PublicID:         sec.ID,
		ISIN:             sec.ISIN,
		Name:             sec.Name,
		SecurityType:     sec.SecurityType,
		MaturityDate:     maturityDate,
		NominalValue:     sec.NominalValue,
		Currency:         sec.Currency,
		CouponRate:       0,
		CouponPeriodDays: 0,
		DayCountBasis:    0,
	}, nil
}

// securityAdapter implements the gateway securities port on top of the
// in-process securities module.
type securityAdapter struct {
	module *securitiesapi.Module
}

var _ ports.SecurityService = securityAdapter{}

func (s securityAdapter) GetSecuritiesByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.ListSecuritiesResponse, error) {
	params, err := toSecurityParams(req)
	if err != nil {
		return dtos.ListSecuritiesResponse{}, status.Error(codes.InvalidArgument, "invalid filter")
	}

	list, err := s.module.List(ctx, params)
	if err != nil {
		return dtos.ListSecuritiesResponse{}, mapSecurityError(err)
	}
	return dtos.ListSecuritiesResponse{Securities: toSecurityDTOs(list)}, nil
}

func (s securityAdapter) GetByParams(ctx context.Context, req dtos.ListSecuritiesRequest) (dtos.Security, error) {
	params, err := toSecurityGetParams(req)
	if err != nil {
		return dtos.Security{}, status.Error(codes.InvalidArgument, "invalid filter")
	}

	sec, err := s.module.Get(ctx, params)
	if err != nil {
		return dtos.Security{}, mapSecurityError(err)
	}
	if sec == nil {
		return dtos.Security{}, status.Error(codes.NotFound, "security not found")
	}
	return toSecurityDTO(*sec), nil
}

func (s securityAdapter) GetEnableSecurities(ctx context.Context) (dtos.ListSecuritiesResponse, error) {
	list, err := s.module.GetEnabled(ctx)
	if err != nil {
		return dtos.ListSecuritiesResponse{}, mapSecurityError(err)
	}
	return dtos.ListSecuritiesResponse{Securities: toSecurityDTOs(list)}, nil
}

func (s securityAdapter) Close() error { return nil }

func mapSecurityError(err error) error {
	switch {
	case errors.Is(err, securitiesapi.ErrSecurityNotFound):
		return status.Error(codes.NotFound, "security not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func toSecurityParams(req dtos.ListSecuritiesRequest) (securitiesapi.Params, error) {
	params := securitiesapi.Params{
		Serie:         appendNonEmpty(req.Series, req.Serie),
		SecurityTypes: appendNonEmpty(req.SecurityTypes, req.SecurityType),
		Name:          req.Name,
		Query:         firstNonEmpty(req.Q, req.Query),
		ISIN:          req.Isin,
	}

	rawIDs := appendNonEmpty(req.Ids, req.Id)
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return securitiesapi.Params{}, err
		}
		params.IDs = append(params.IDs, id)
	}
	params.ISINs = appendNonEmpty(req.Isins, req.Isin)

	var err error
	if params.IssueAfter, err = parseDate(req.IssueAfter); err != nil {
		return securitiesapi.Params{}, err
	}
	if params.IssueBefore, err = parseDate(req.IssueBefore); err != nil {
		return securitiesapi.Params{}, err
	}
	if params.MaturityAfter, err = parseDate(req.MaturityAfter); err != nil {
		return securitiesapi.Params{}, err
	}
	if params.MaturityBefore, err = parseDate(req.MaturityBefore); err != nil {
		return securitiesapi.Params{}, err
	}
	return params, nil
}

func toSecurityGetParams(req dtos.ListSecuritiesRequest) (securitiesapi.GetParams, error) {
	params := securitiesapi.GetParams{
		ID:           req.Id,
		ISIN:         req.Isin,
		Serie:        req.Serie,
		SecurityType: req.SecurityType,
	}

	var err error
	if params.IssueAfter, err = parseDate(req.IssueAfter); err != nil {
		return securitiesapi.GetParams{}, err
	}
	if params.IssueBefore, err = parseDate(req.IssueBefore); err != nil {
		return securitiesapi.GetParams{}, err
	}
	if params.MaturityAfter, err = parseDate(req.MaturityAfter); err != nil {
		return securitiesapi.GetParams{}, err
	}
	if params.MaturityBefore, err = parseDate(req.MaturityBefore); err != nil {
		return securitiesapi.GetParams{}, err
	}
	return params, nil
}

func parseDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}

func appendNonEmpty(values []string, extra string) []string {
	out := make([]string, 0, len(values)+1)
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	if extra != "" {
		out = append(out, extra)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func toSecurityDTOs(list []securitiesapi.Security) []dtos.Security {
	out := make([]dtos.Security, len(list))
	for i, s := range list {
		out[i] = toSecurityDTO(s)
	}
	return out
}

func toSecurityDTO(s securitiesapi.Security) dtos.Security {
	nominal, _ := strconv.ParseFloat(s.NominalValue, 64)
	fx, _ := strconv.ParseFloat(s.Fx, 64)
	return dtos.Security{
		ID:           s.ID,
		Isin:         s.ISIN,
		Name:         s.Name,
		SecurityType: s.SecurityType,
		Serie:        s.Serie,
		IssueDate:    s.IssueDate,
		MaturityDate: s.MaturityDate,
		NominalValue: nominal,
		Currency:     s.Currency,
		Fx:           fx,
		Enable:       s.Enable,
	}
}

// ordersAdapter implements the gateway orders port on top of the in-process
// orders module.
type ordersAdapter struct {
	module *ordersapi.Module
}

var _ ports.OrdersService = ordersAdapter{}

func (o ordersAdapter) CreateOrder(ctx context.Context, req dtos.CreateOrderRequest) (dtos.OrderResponse, error) {
	res, err := o.module.CreateOrder(ctx, ordersapi.CreateOrderCommand{
		TraderID:   req.TraderID,
		SecurityID: req.SecurityID,
		Direction:  req.Direction,
		OrderType:  req.OrderType,
		Price:      req.Price,
		Yield:      req.Yield,
		Quantity:   req.Quantity,
		Expiration: req.Expiration.Value(),
	})
	if err != nil {
		return dtos.OrderResponse{}, mapOrderError(err)
	}
	return toOrderDTO(res), nil
}

func (o ordersAdapter) CancelOrder(ctx context.Context, req dtos.CancelOrderRequest) (dtos.OrderResponse, error) {
	res, err := o.module.CancelOrder(ctx, ordersapi.CancelOrderCommand{
		OrderID:  req.OrderID,
		TraderID: req.TraderID,
	})
	if err != nil {
		return dtos.OrderResponse{}, mapOrderError(err)
	}
	return toOrderDTO(res), nil
}

func (o ordersAdapter) ListOrders(ctx context.Context, req dtos.ListOrdersRequest) (dtos.ListOrdersResponse, error) {
	list, err := o.module.ListOrders(ctx, ordersapi.ListOrdersQuery{
		TraderID:  req.TraderID,
		OrderIDs:  req.OrderIDs,
		Direction: req.Direction,
		Status:    req.Status,
	})
	if err != nil {
		return dtos.ListOrdersResponse{}, mapOrderError(err)
	}
	return dtos.ListOrdersResponse{Orders: toOrderDTOs(list)}, nil
}

func (o ordersAdapter) GetActiveOrders(ctx context.Context, req dtos.GetActiveOrdersRequest) (dtos.GetActiveOrdersResponse, error) {
	list, err := o.module.ActiveOrders(ctx, ordersapi.ActiveOrdersQuery{
		TraderID:   req.TraderID,
		SecurityID: req.SecurityID,
		Direction:  req.Direction,
	})
	if err != nil {
		return dtos.GetActiveOrdersResponse{}, mapOrderError(err)
	}
	return dtos.GetActiveOrdersResponse{Orders: toOrderDTOs(list)}, nil
}

func (o ordersAdapter) GetMarketDepth(ctx context.Context, req dtos.GetMarketDepthRequest) (dtos.MarketDepthResponse, error) {
	securityID, err := uuid.Parse(req.SecurityID)
	if err != nil {
		return dtos.MarketDepthResponse{}, status.Error(codes.InvalidArgument, "invalid security_id")
	}

	depth, err := o.module.MarketDepth(ctx, securityID)
	if err != nil {
		return dtos.MarketDepthResponse{}, mapOrderError(err)
	}
	return dtos.MarketDepthResponse{
		Bids: toDepthDTOs(depth.Bids),
		Asks: toDepthDTOs(depth.Asks),
	}, nil
}

func (o ordersAdapter) Close() error { return nil }

func mapOrderError(err error) error {
	switch {
	case errors.Is(err, ordersapi.ErrOrderNotFound),
		errors.Is(err, ordersapi.ErrTradeNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ordersapi.ErrOrderAlreadyCanceled),
		errors.Is(err, ordersapi.ErrCannotCancelMarketOrder):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, ordersapi.ErrInvalidTraderID),
		errors.Is(err, ordersapi.ErrInvalidSecurityID),
		errors.Is(err, ordersapi.ErrInvalidPrice),
		errors.Is(err, ordersapi.ErrInvalidYield),
		errors.Is(err, ordersapi.ErrInvalidQuantity),
		errors.Is(err, ordersapi.ErrInvalidNominalValue),
		errors.Is(err, ordersapi.ErrPriceInMarketOrder),
		errors.Is(err, ordersapi.ErrQtyTooLarge),
		errors.Is(err, ordersapi.ErrYieldInMarketOrder),
		errors.Is(err, ordersapi.ErrInvalidFX),
		errors.Is(err, ordersapi.ErrInvalidDirection),
		errors.Is(err, ordersapi.ErrInvalidOrderType),
		errors.Is(err, ordersapi.ErrInvalidOrderStatus),
		errors.Is(err, ordersapi.ErrInvalidOrderId),
		errors.Is(err, ordersapi.ErrInvalidExpiration),
		errors.Is(err, ordersapi.ErrMarketOrderExpiration),
		errors.Is(err, ordersapi.ErrInvalidTakerID),
		errors.Is(err, ordersapi.ErrInvalidMakerID):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func toOrderDTO(o ordersapi.Order) dtos.OrderResponse {
	return dtos.OrderResponse{
		ID:                o.ID,
		SecurityID:        o.SecurityID,
		Direction:         o.Direction,
		OrderType:         o.OrderType,
		Price:             stringPtr(o.Price),
		Yield:             stringPtr(o.Yield),
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

func toOrderDTOs(list []ordersapi.Order) []dtos.OrderResponse {
	out := make([]dtos.OrderResponse, len(list))
	for i, o := range list {
		out[i] = toOrderDTO(o)
	}
	return out
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func toDepthDTOs(levels []ordersapi.MarketDepthLevel) []*dtos.MarketDepthLevel {
	out := make([]*dtos.MarketDepthLevel, 0, len(levels))
	for _, level := range levels {
		out = append(out, &dtos.MarketDepthLevel{
			Price:        level.Price,
			NominalValue: level.NominalValue,
			Quantity:     level.Quantity,
			OrderCounts:  level.OrderCounts,
			Direction:    level.Direction,
		})
	}
	return out
}
