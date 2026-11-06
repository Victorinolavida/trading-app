package securityhandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
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
	cl, err := authmiddleware.GetClaims(ctx)
	if err == nil {
		ctx = grpcclient.WithUserID(ctx, cl.UserID)
	}
	return ctx
}

var ErrMissingValidator = errors.New("security handler: validator is required")

type SecurityHandler struct {
	service    ports.SecurityService
	validate   *validator.Validate
	jwtService *jwt.TokenService
	cfg        *config.Config
}

var _ module.Module = (*SecurityHandler)(nil)

func NewSecurityHandler(cfg *config.Config, service ports.SecurityService, validate *validator.Validate) (*SecurityHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	jwtService, err := jwt.NewService(cfg.Token)
	if err != nil {
		return nil, err
	}
	return &SecurityHandler{service: service, validate: validate, jwtService: jwtService, cfg: cfg}, nil
}

func (h *SecurityHandler) BasePath() string {
	return "/securities"
}

func (h *SecurityHandler) List(w http.ResponseWriter, r *http.Request) {
	req := dtos.ListSecuritiesRequest{
		Id:             strings.TrimSpace(r.URL.Query().Get("id")),
		Isin:           strings.TrimSpace(r.URL.Query().Get("isin")),
		Name:           strings.TrimSpace(r.URL.Query().Get("name")),
		Query:          strings.TrimSpace(r.URL.Query().Get("q")),
		Q:              strings.TrimSpace(r.URL.Query().Get("q")),
		Currency:       strings.TrimSpace(r.URL.Query().Get("currency")),
		Serie:          strings.TrimSpace(r.URL.Query().Get("serie")),
		SecurityType:   strings.TrimSpace(r.URL.Query().Get("security_type")),
		IssueBefore:    strings.TrimSpace(r.URL.Query().Get("issue_before")),
		IssueAfter:     strings.TrimSpace(r.URL.Query().Get("issue_after")),
		MaturityBefore: strings.TrimSpace(r.URL.Query().Get("maturity_before")),
		MaturityAfter:  strings.TrimSpace(r.URL.Query().Get("maturity_after")),
	}

	if ids := r.URL.Query()["ids"]; len(ids) > 0 {
		req.Ids = ids
	}
	if isins := r.URL.Query()["isins"]; len(isins) > 0 {
		req.Isins = isins
	}
	if series := r.URL.Query()["series"]; len(series) > 0 {
		req.Series = series
	}
	if types := r.URL.Query()["security_types"]; len(types) > 0 {
		req.SecurityTypes = types
	}

	result, err := h.service.GetSecuritiesByParams(withRequestMeta(r.Context(), r), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, result, http.StatusOK); err != nil {
		slog.Error("error writing response", slog.Any("error", err))
	}
}

func (h *SecurityHandler) Get(w http.ResponseWriter, r *http.Request) {
	req := dtos.ListSecuritiesRequest{
		Id:             strings.TrimSpace(r.URL.Query().Get("id")),
		Isin:           strings.TrimSpace(r.URL.Query().Get("isin")),
		Name:           strings.TrimSpace(r.URL.Query().Get("name")),
		Query:          strings.TrimSpace(r.URL.Query().Get("q")),
		Q:              strings.TrimSpace(r.URL.Query().Get("q")),
		Currency:       strings.TrimSpace(r.URL.Query().Get("currency")),
		Serie:          strings.TrimSpace(r.URL.Query().Get("serie")),
		SecurityType:   strings.TrimSpace(r.URL.Query().Get("security_type")),
		IssueBefore:    strings.TrimSpace(r.URL.Query().Get("issue_before")),
		IssueAfter:     strings.TrimSpace(r.URL.Query().Get("issue_after")),
		MaturityBefore: strings.TrimSpace(r.URL.Query().Get("maturity_before")),
		MaturityAfter:  strings.TrimSpace(r.URL.Query().Get("maturity_after")),
	}

	result, err := h.service.GetByParams(withRequestMeta(r.Context(), r), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, result, http.StatusOK); err != nil {
		slog.Error("error writing response", slog.Any("error", err))
	}
}

func (h *SecurityHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetEnableSecurities(withRequestMeta(r.Context(), r))
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	if err := response.WriteJsonResponse(w, result, http.StatusOK); err != nil {
		slog.Error("error writing response", slog.Any("error", err))
	}
}

func (h *SecurityHandler) Routes() http.Handler {
	r := chi.NewMux()
	protected := r.With(authmiddleware.AuthMiddleware(h.jwtService, &h.cfg.Cookie))
	protected.Get("/", h.List)
	protected.Get("/search", h.Get)
	protected.Get("/active", h.GetActive)
	protected.Get("/enabled", h.GetActive)
	return r
}

