package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/stream"
)

type stubModule struct {
	basePath string
}

func TestShutdownClosesStreamsAndWaitsForOtherRequests(t *testing.T) {
	h := NewHttpServer(config.Config{Env: "development"}).(*httpServer)
	streamStarted := make(chan struct{})
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	h.Handle("/stream", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(*response.RequestWriter).ResponseWriter.(http.Flusher).Flush()
		close(streamStarted)
		<-stream.Done(r.Context())
	})
	h.Handle("/work", func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		select {
		case <-releaseRequest:
			w.WriteHeader(http.StatusNoContent)
		case <-r.Context().Done():
			t.Error("ordinary request was canceled")
		}
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = h.srv.Serve(listener) }()
	url := "http://" + listener.Addr().String()
	response, err := http.Get(url + "/stream")
	require.NoError(t, err)
	defer response.Body.Close()
	<-streamStarted
	workDone := make(chan int, 1)
	go func() {
		r, err := http.Get(url + "/work")
		if err != nil {
			workDone <- 0
			return
		}
		defer r.Body.Close()
		workDone <- r.StatusCode
	}()
	<-requestStarted
	shutdownDone := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go func() { shutdownDone <- h.Shutdown(ctx) }()
	select {
	case err := <-shutdownDone:
		t.Fatalf("shutdown returned before ordinary request completed: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseRequest)
	select {
	case code := <-workDone:
		require.Equal(t, http.StatusNoContent, code)
	case <-ctx.Done():
		t.Fatal("request did not complete")
	}
	select {
	case err := <-shutdownDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal("shutdown did not close stream")
	}
}

func TestShutdownClosesUnresponsiveConnectionsAfterDeadline(t *testing.T) {
	h := NewHttpServer(config.Config{Env: "development"}).(*httpServer)
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	h.Handle("/stuck", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(*response.RequestWriter).ResponseWriter.(http.Flusher).Flush()
		close(started)
		<-release
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = h.srv.Serve(listener) }()
	res, err := http.Get("http://" + listener.Addr().String() + "/stuck")
	require.NoError(t, err)
	defer res.Body.Close()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	err = h.Shutdown(ctx)
	require.True(t, errors.Is(err, context.DeadlineExceeded), "expected graceful deadline, got %v", err)
	readDone := make(chan error, 1)
	go func() { var b [1]byte; _, err := res.Body.Read(b[:]); readDone <- err }()
	select {
	case err := <-readDone:
		require.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("connection remained open")
	}
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
