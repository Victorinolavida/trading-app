package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
)

const (
	defaultPort  = 8080
	drainTimeout = 30 * time.Second
)

type ServerInterface interface {
	Run(context.Context) error
	Start() error
	Shutdown(context.Context) error
	ShutdownSignal() <-chan struct{}
	Mux() *chi.Mux
	Mount(pattern string, handle http.Handler)
	MountModule(prefix string, m module.Module)
	Handle(pattern string, handler http.HandlerFunc)
	Method(method, pattern string, handler http.HandlerFunc)
	ChiRoutes() chi.Routes
}

var _ ServerInterface = (*httpServer)(nil)

type httpServer struct {
	srv            *http.Server
	router         *chi.Mux
	shutdownSignal chan struct{}
	shutdownOnce   sync.Once
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
		// Server sent events streams hold the connection open indefinitely,
		// so the idle timeout has to stay disabled.
		IdleTimeout: 0,
	}

	httpSvc := &httpServer{
		srv:            srv,
		router:         r,
		shutdownSignal: make(chan struct{}),
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

func (h *httpServer) ChiRoutes() chi.Routes {
	return h.router
}

func (h *httpServer) ShutdownSignal() <-chan struct{} {
	return h.shutdownSignal
}

func (h *httpServer) notifyShutdown() {
	h.shutdownOnce.Do(func() {
		close(h.shutdownSignal)
	})
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

func (h *httpServer) Run(ctx context.Context) error {
	defer h.notifyShutdown()

	listener, err := net.Listen("tcp", h.srv.Addr)
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- h.srv.Serve(listener)
	}()

	slog.Info("HTTP server starting", slog.String("addr", listener.Addr().String()))

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received; draining HTTP server", slog.Duration("timeout", drainTimeout))

		h.notifyShutdown()

		drainCtx, cancel := context.WithTimeout(context.Background(), drainTimeout)
		defer cancel()

		shutdownErr := h.srv.Shutdown(drainCtx)
		if shutdownErr != nil {
			slog.Error("graceful HTTP shutdown failed; forcing close", slog.Any("error", shutdownErr))
			_ = h.srv.Close()
		}

		_ = listener.Close()

		serveErr := <-serverErr
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) && !errors.Is(serveErr, net.ErrClosed) {
			return serveErr
		}
		return shutdownErr
	}
}

func (h *httpServer) Start() error {
	slog.Info("HTTP server starting", slog.String("addr", h.srv.Addr))
	return h.srv.ListenAndServe()
}

func (h *httpServer) Shutdown(ctx context.Context) error {
	slog.Info("HTTP server shutting down...")
	h.notifyShutdown()
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
