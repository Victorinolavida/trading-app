package ordershandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
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
	protected.Post("/cancel", h.CancelOrder)
	protected.Get("/{orderGroupId}", h.GetOrderGroup)
	return r
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RenderBadRequest(w, "invalid body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderBadRequest(w, err.Error())
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
	var req dtos.CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RenderBadRequest(w, "invalid body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}

	if claims, err := authmiddleware.GetClaims(r.Context()); err == nil {
		if req.TraderID == "" {
			req.TraderID = claims.UserID
		}
	}

	res, err := h.ordersService.CancelOrder(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *OrdersHandler) GetOrderGroup(w http.ResponseWriter, r *http.Request) {
	orderGroupID := chi.URLParam(r, "orderGroupId")
	res, err := h.ordersService.GetOrderGroup(r.Context(), orderGroupID)
	if err != nil {
		response.RenderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *OrdersHandler) BasePath() string {
	return "/orders"
}
