package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/config"
)

type stubModule struct {
	basePath string
}

func (m stubModule) BasePath() string { return m.basePath }

func (m stubModule) Routes() http.Handler {
	r := chi.NewMux()
	r.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return r
}

func TestNewHttpServer_ReturnsInterface(t *testing.T) {
	srv := NewHttpServer(config.Config{Env: "development"})

	require.NotNil(t, srv)
	assert.NotNil(t, srv.Mux())
}

func TestHttpServer_Health(t *testing.T) {
	srv := NewHttpServer(config.Config{Env: "development"})

	rr := httptest.NewRecorder()
	srv.Mux().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
}

func TestHttpServer_MountModule(t *testing.T) {
	srv := NewHttpServer(config.Config{Env: "development"})

	srv.MountModule("/api/v0", stubModule{basePath: "/auth"})

	rr := httptest.NewRecorder()
	srv.Mux().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v0/auth/ping", nil))

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestHttpServer_NotFound(t *testing.T) {
	srv := NewHttpServer(config.Config{Env: "development"})

	rr := httptest.NewRecorder()
	srv.Mux().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/nope", nil))

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
