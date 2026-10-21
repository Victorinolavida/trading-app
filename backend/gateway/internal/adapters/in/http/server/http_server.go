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
)

var defaultPort = 8080

type ServerInterface interface {
	Start() error
	Shutdown(context.Context) error
	Mux() *chi.Mux
	Mount(pattern string, handle http.Handler)
	Handle(pattern string, handler http.HandlerFunc)
	Method(method, pattern string, handler http.HandlerFunc)
}

type Module interface {
	BasePath() string
	Routes() http.Handler
}

var _ ServerInterface = (*httpServer)(nil)

type httpServer struct {
	srv    *http.Server
	router *chi.Mux
}

func NewHttpServer(conf config.Config) *httpServer {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromHeader("X-Real-IP"))
	r.Use(middleware.Recoverer)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		RenderNotFound(w)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		RenderMethodNotAllowed(w)
	})

	allowOrigins := []string{"*"}
	if conf.IsProduction() {
		allowOrigins = conf.Server.AllowOrigins
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
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
		WriteJsonResponse(w, Envelop("message", "ok"), http.StatusOK)
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
