package authhandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	"github.com/victorinolavida/pkg/jwt"
)

func withRequestMeta(ctx context.Context, r *http.Request) context.Context {
	ctx = grpcclient.WithClientIP(ctx, chimw.GetClientIP(r.Context()))
	ctx = grpcclient.WithUserAgent(ctx, r.UserAgent())
	cl, err := middleware.GetClaims(ctx)
	if err == nil {
		ctx = grpcclient.WithUserID(ctx, cl.UserID)
	}
	return ctx
}

// ErrMissingValidator is returned when the handler is built without a validator.
var ErrMissingValidator = errors.New("auth handler: validator is required")

type AuthHandler struct {
	service    ports.AuthService
	validate   *validator.Validate
	jwtService *jwt.TokenService
}

var _ module.Module = (*AuthHandler)(nil)

func NewAuthHandler(conf *config.Config, service ports.AuthService, validate *validator.Validate) (*AuthHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	jwtService, err := jwt.NewService(conf.Token)
	if err != nil {
		return nil, err
	}
	return &AuthHandler{service, validate, jwtService}, nil
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
	result, err := h.service.Login(withRequestMeta(r.Context(), r), dtos.LoginRequest{
		Email:    loginReq.Email,
		Password: loginReq.Password,
	})
	if err != nil {
		response.RenderGRPCError(w, r, err)
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
	result, err := h.service.Register(withRequestMeta(r.Context(), r), dtos.RegisterRequest{
		Email:     registerReq.Email,
		FirstName: registerReq.FirstName,
		LastName:  registerReq.LastName,
		Password:  registerReq.Password,
	})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusCreated); err != nil {
		slog.Error("error writing register response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Whoami(w http.ResponseWriter, r *http.Request) {
	token, _ := response.GetTokenFromRequest(r)

	if token == "" {
		response.RenderUnauthorized(w)
		return
	}

	// send message via grpc
	result, err := h.service.Whoami(withRequestMeta(r.Context(), r), dtos.WhoamiUserRequest{
		Token: token,
	})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, newWhoamiResponse(result), http.StatusOK); err != nil {
		slog.Error("error writing whoami response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewMux()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)

	protected := r.With(middleware.AuthMiddleware(h.jwtService))
	protected.Get("/whoami", h.Whoami)
	return r
}
