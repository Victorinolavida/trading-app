package authhandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
)

// ErrMissingValidator is returned when the handler is built without a validator.
var ErrMissingValidator = errors.New("auth handler: validator is required")

type AuthHandler struct {
	service  ports.AuthService
	validate *validator.Validate
}

var _ module.Module = (*AuthHandler)(nil)

func NewAuthHandler(service ports.AuthService, validate *validator.Validate) (*AuthHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	return &AuthHandler{service, validate}, nil
}

func (h *AuthHandler) BasePath() string {
	return "/auth"
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginReq, err := response.ParseBodyFromRequest[LoginRequest](w, r)
	if err != nil {
		slog.Error("error parsing body", slog.Any("error", err))
		response.RenderBadRequest(w, "can not parse the body request")
		return
	}

	if err := h.validate.Struct(loginReq); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	// send message via grpc
	result, err := h.service.Login(r.Context(), dtos.LoginRequest{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	})
	if err != nil {
		response.RenderGRPCError(w, err)
		return
	}

	if err := response.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusOK); err != nil {
		slog.Error("error writing login response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	registerReq, err := response.ParseBodyFromRequest[RegisterRequest](w, r)
	if err != nil {
		slog.Error("error parsing body", slog.Any("error", err))
		response.RenderBadRequest(w, "can not parse the body request")
		return
	}

	if err := h.validate.Struct(registerReq); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	// send message via grpc
	result, err := h.service.Register(r.Context(), dtos.RegisterRequest{
		Email:     registerReq.Email,
		FirstName: registerReq.FirstName,
		LastName:  registerReq.LastName,
		Password:  registerReq.Password,
	})
	if err != nil {
		response.RenderGRPCError(w, err)
		return
	}

	if err := response.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusCreated); err != nil {
		slog.Error("error writing register response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewMux()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	return r
}
