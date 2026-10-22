package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
)

var defaultPort = 8080

type ServerInterface interface {
	Start() error
	Shutdown(context.Context) error
	Mux() *chi.Mux
	Mount(pattern string, handle http.Handler)
	MountModule(prefix string, m module.Module)
	Handle(pattern string, handler http.HandlerFunc)
	Method(method, pattern string, handler http.HandlerFunc)
}

var _ ServerInterface = (*httpServer)(nil)

type httpServer struct {
	srv    *http.Server
	router *chi.Mux
}

func NewHttpServer(conf config.Config) ServerInterface {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(response.RequestIDResponseMiddleware)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				w.Header().Set(middleware.RequestIDHeader, reqID)
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.ClientIPFromHeader("X-Real-IP"))
	r.Use(middleware.Recoverer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.RenderNotFound(w)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.RenderMethodNotAllowed(w)
	})

	// The browser session is cookie based, so the cors handler has to allow
	// credentials. Browsers reject "*" together with credentials, therefore the
	// allowed origins have to be listed explicitly.
	allowOrigins := conf.Server.AllowOrigins
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"*"}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of the major browsers
	}))

	srv := &http.Server{
		Addr:        formatServerAddress(conf.Server.Host, conf.Server.Port),
		Handler:     r,
		ReadTimeout: time.Duration(10 * time.Second),
		IdleTimeout: time.Duration(100 * time.Second),
	}

	httpSvc := &httpServer{
		srv:    srv,
		router: r,
	}

	httpSvc.Method(http.MethodGet, "/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJsonResponse(w, response.Envelop("message", "ok"), http.StatusOK)
	})

	return httpSvc
}

// Mux returns the underlying chi router
func (h *httpServer) Mux() *chi.Mux {
	return h.router
}

// Method registers and HTTP handler for a specific method and pattern
func (h *httpServer) Method(method string, pattern string, handler http.HandlerFunc) {
	h.router.MethodFunc(method, pattern, handler)
}

// Handle register an HTTP handler for all method on a specific pattern
func (h *httpServer) Handle(pattern string, handler http.HandlerFunc) {
	h.router.Handle(pattern, handler)
}

// Mount mounts a sub-router (or another http.Handler) at specific path
func (h *httpServer) Mount(pattern string, handler http.Handler) {
	h.router.Mount(pattern, handler)
}

// MountModule mounts a feature module under its base path, prefixed with the
// given api prefix, e.g. "/api/v0" + "/auth".
func (h *httpServer) MountModule(prefix string, m module.Module) {
	h.router.Mount(prefix+m.BasePath(), m.Routes())
}

func (h *httpServer) Start() error {
	slog.Info("HTTP server starting", slog.String("addr", h.srv.Addr))
	return h.srv.ListenAndServe()
}

func (h *httpServer) Shutdown(ctx context.Context) error {
	slog.Info("HTTP server shutting down...")
	return h.srv.Shutdown(ctx)
}

func formatServerAddress(host string, port int) string {
	if port == 0 {
		port = defaultPort
	}
	if host == "" {
		host = "localhost"
	}
	return host + ":" + strconv.Itoa(port)
}
