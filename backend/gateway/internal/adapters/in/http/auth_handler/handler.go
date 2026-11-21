package authhandler

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/cookies"
	authmiddleware "github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/passwordreset"
	"github.com/victorinolavida/gateway/internal/application/ports"
	grpcclient "github.com/victorinolavida/pkg/grpc_client"
	"github.com/victorinolavida/pkg/jwt"
)

func withRequestMeta(ctx context.Context, r *http.Request) context.Context {
	ctx = grpcclient.WithClientIP(ctx, chimw.GetClientIP(r.Context()))
	ctx = grpcclient.WithUserAgent(ctx, r.UserAgent())
	cl, err := authmiddleware.GetClaims(ctx)
	if err == nil {
		ctx = grpcclient.WithUserID(ctx, cl.UserID)
	}
	return ctx
}

// ErrMissingValidator is returned when the handler is built without a validator.
var ErrMissingValidator = errors.New("auth handler: validator is required")

type AuthHandler struct {
	service      ports.AuthService
	validate     *validator.Validate
	jwtService   ports.TokenVerifier
	cfg          *config.Config
	resetLimiter *passwordreset.Limiter
}

func (h *AuthHandler) SetTokenVerifier(v ports.TokenVerifier) { h.jwtService = v }

var _ module.Module = (*AuthHandler)(nil)

func NewAuthHandler(cfg *config.Config, service ports.AuthService, validate *validator.Validate) (*AuthHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	jwtService, err := jwt.NewService(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &AuthHandler{service: service, validate: validate, jwtService: jwtService, cfg: cfg, resetLimiter: passwordreset.NewLimiter()}, nil
}

func clientAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[PasswordResetRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, "invalid request")
		return
	}
	if err = h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}
	if !h.resetLimiter.Allow(clientAddress(r), req.Email) {
		response.RenderError(w, http.StatusTooManyRequests, "too many requests")
		return
	}
	if err = h.service.RequestPasswordReset(withRequestMeta(r.Context(), r), req.Email); err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[PasswordResetConfirmation](w, r)
	if err != nil {
		response.RenderBadRequest(w, "invalid request")
		return
	}
	if err = h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}
	ip := clientAddress(r)
	if !h.resetLimiter.Allow(ip, "confirm:"+ip) {
		response.RenderError(w, http.StatusTooManyRequests, "too many requests")
		return
	}
	if err = h.service.ConfirmPasswordReset(withRequestMeta(r.Context(), r), req.Token, req.NewPassword); err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) BasePath() string {
	return "/auth"
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	loginReq, err := response.ParseBodyFromRequest[LoginRequest](w, r)
	if err != nil {
		slog.Error("error parsing body", slog.Any("error", err))
		response.RenderBadRequest(w, err.Error())
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

	if err := cookies.Set(w, &h.cfg.Cookie, result.Token); err != nil {
		slog.Error("error setting session cookies", slog.Any("error", err))
		response.RenderServerError(w, err)
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
		response.RenderBadRequest(w, err.Error())
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

	if err := cookies.Set(w, &h.cfg.Cookie, result.Token); err != nil {
		slog.Error("error setting session cookies", slog.Any("error", err))
		response.RenderServerError(w, err)
		return
	}

	if err := response.WriteJsonResponse(w, newAuthUserResponse(result), http.StatusCreated); err != nil {
		slog.Error("error writing register response", slog.Any("error", err))
	}
}

func (h *AuthHandler) Whoami(w http.ResponseWriter, r *http.Request) {
	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		response.RenderUnauthorized(w)
		return
	}
	if claims.UserID == "" {
		response.RenderUnauthorized(w)
		return
	}
	// send message via grpc
	result, err := h.service.Whoami(withRequestMeta(r.Context(), r), dtos.WhoamiUserRequest{
		UserID: claims.UserID,
	})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, newWhoamiResponse(result), http.StatusOK); err != nil {
		slog.Error("error writing whoami response", slog.Any("error", err))
	}
}

// Logout clears the session cookies.
//
// It is registered outside the protected group on purpose: an expired token
// must still be able to clear its own cookies, otherwise the browser keeps
// sending a credential the server will always reject.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookies.Clear(w, &h.cfg.Cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewMux()
	r.Post("/login", h.Login)
	r.Post("/register", h.Register)
	r.Post("/logout", h.Logout)
	r.Post("/password-reset/request", h.RequestPasswordReset)
	r.Post("/password-reset/confirm", h.ConfirmPasswordReset)

	protected := r.With(authmiddleware.AuthMiddleware(h.jwtService, &h.cfg.Cookie))
	protected.Get("/whoami", h.Whoami)
	return r
}
