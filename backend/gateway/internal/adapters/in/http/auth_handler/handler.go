package authhandler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	httpserver "github.com/victorinolavida/gateway/internal/adapters/in/http/server"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	"github.com/victorinolavida/pkg/logger"
)

type AuthHandler struct {
	service  ports.AuthService
	validate *validator.Validate
}

var _ httpserver.Module = (*AuthHandler)(nil)

func NewAuthHandler(service ports.AuthService, validate *validator.Validate) *AuthHandler {
	if validate == nil {
		logger.Fatal("validator dependency is required")
	}
	return &AuthHandler{service, validate}
}

func (h *AuthHandler) BasePath() string {
	return "/auth"
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginData, err := httpserver.ParseBodyFromRequest[dtos.LoginRequest](w, r)
	if err != nil {
		slog.Error("error parsing body", slog.Any("error", err))
		httpserver.RenderBadRequest(w, "can not parse the body request")
		return
	}

	if err := h.validate.Struct(loginData); err != nil {
		httpserver.RenderBadRequest(w, err.Error())
		return
	}

	// send message via grpc
	result, err := h.service.Login(r.Context(), loginData)
	if err != nil {
		httpserver.RenderServerError(w, err)
		return
	}

	if err := httpserver.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusOK); err != nil {
		slog.Error("error writing login response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	registerData, err := httpserver.ParseBodyFromRequest[dtos.RegisterRequest](w, r)
	if err != nil {
		slog.Error("error parsing body", slog.Any("error", err))
		httpserver.RenderBadRequest(w, "can not parse the body request")
		return
	}

	if err := h.validate.Struct(registerData); err != nil {
		httpserver.RenderBadRequest(w, err.Error())
		return
	}

	// send message via grpc
	result, err := h.service.Register(r.Context(), registerData)
	if err != nil {
		httpserver.RenderServerError(w, err)
		return
	}

	if err := httpserver.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusCreated); err != nil {
		slog.Error("error writing register response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewMux()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	return r
}
