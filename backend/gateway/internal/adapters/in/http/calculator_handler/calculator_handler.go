package calculatorhandler

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	calculatorapi "github.com/victorinolavida/calculator/api"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/pkg/jwt"
)

var ErrMissingValidator = errors.New("calculator handler: validator is required")

type CalculatorHandler struct {
	calculator calculatorapi.Module
	validate   *validator.Validate
	jwtService *jwt.TokenService
}

func (h *CalculatorHandler) CalculatePrice(ctx context.Context, req calculatorapi.PriceRequest) (calculatorapi.PriceResult, error) {
	return h.calculator.Price(ctx, req)
}

func (h *CalculatorHandler) CalculateYield(ctx context.Context, req calculatorapi.YieldRequest) (float64, error) {
	return h.calculator.Yield(ctx, req)
}

func NewCalculatorHandler(calc *calculatorapi.Module, validate *validator.Validate, jwtService *jwt.TokenService) *CalculatorHandler {
	return &CalculatorHandler{
		calculator: *calc,
		validate:   validate,
		jwtService: jwtService,
	}
}

var _ module.Module = (*CalculatorHandler)(nil)

func (h *CalculatorHandler) Routes() http.Handler {
	r := chi.NewMux()
	protected := r.With(authmiddleware.AuthMiddleware(h.jwtService, nil))
	protected.Post("/price", h.Price)
	protected.Post("/yield", h.Yield)
	protected.Post("/cashflows", h.Cashflows)
	return r
}

func (h *CalculatorHandler) Price(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[calculatorapi.PriceRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	res, err := h.calculator.Price(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *CalculatorHandler) Yield(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[calculatorapi.YieldRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

 rate, err := h.calculator.Yield(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, rate, http.StatusOK)
}

func (h *CalculatorHandler) Cashflows(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[calculatorapi.CashflowRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	cfs, err := h.calculator.Cashflows(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	result := make([]dtos.CashFlow, 0, len(cfs))
	for _, cf := range cfs {
		result = append(result, dtos.CashFlow{
			Date: cf.Date.String(),
			Rate: cf.Amount.String(),
		})
	}
	response.RenderJson(w, result, http.StatusOK)
}

func (h *CalculatorHandler) BasePath() string {
	return "/calculator"
}