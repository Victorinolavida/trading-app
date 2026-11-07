package ordershandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/config"
	authmiddleware "github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/jwt"
)

var ErrMissingValidator = errors.New("orders handler: validator is required")

type OrdersHandler struct {
	conf          *config.Config
	ordersService ports.OrdersService
	validate      *validator.Validate
	jwtService    *jwt.TokenService
}

func NewOrdersHandler(conf *config.Config, ordersService ports.OrdersService, validate *validator.Validate) (*OrdersHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	jwtService, err := jwt.NewService(conf.Token)
	if err != nil {
		return nil, err
	}
	return &OrdersHandler{
		conf:          conf,
		ordersService: ordersService,
		validate:      validate,
		jwtService:    jwtService,
	}, nil
}

var _ module.Module = (*OrdersHandler)(nil)

func (h *OrdersHandler) Routes() http.Handler {
	r := chi.NewMux()
	protected := r.With(authmiddleware.AuthMiddleware(h.jwtService, &h.conf.Cookie))
	protected.Post("/", h.CreateOrder)
	protected.Patch("/{orderId}/cancel", h.CancelOrder)
	protected.Get("/active", h.GetActiveOrders)
	protected.Get("/", h.ListOrders)
	return r
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[dtos.CreateOrderRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	// Get trader from auth context if available
	if claims, err := authmiddleware.GetClaims(r.Context()); err == nil {
		if req.TraderID == "" {
			req.TraderID = claims.UserID
		}
	}

	res, err := h.ordersService.CreateOrder(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusCreated)
}

func (h *OrdersHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("error getting claims", slog.Any("error", err))
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}

	orderID := chi.URLParam(r, "orderId")
	req := dtos.CancelOrderRequest{
		TraderID: claims.UserID,
		OrderID:  orderID,
	}

	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	res, err := h.ordersService.CancelOrder(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}

	q := r.URL.Query()
	req := dtos.ListOrdersRequest{
		TraderID:  claims.UserID,
		OrderIDs:  parseOrderIDs(q["order_ids"]),
		Direction: q.Get("direction"),
		Status:    q.Get("status"),
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	res, err := h.ordersService.ListOrders(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

// parseOrderIDs accepts repeated ?order_ids=... parameters and
// comma-separated values, dropping empty entries.
func parseOrderIDs(values []string) []string {
	ids := make([]string, 0, len(values))
	for _, v := range values {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func (h *OrdersHandler) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}
	req := dtos.GetActiveOrdersRequest{
		SecurityID: q.Get("security_id"),
		Direction:  q.Get("direction"),
		TraderID:   claims.UserID,
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}
	res, err := h.ordersService.GetActiveOrders(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *OrdersHandler) BasePath() string {
	return "/orders"
}
