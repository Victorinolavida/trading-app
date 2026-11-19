package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/config"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func freePort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func TestHttpServer_RunFailsWhenPortIsAlreadyInUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	srv := NewHttpServer(config.Config{
		Env:    "development",
		Server: config.Server{Host: "127.0.0.1", Port: port},
	})

	require.Error(t, srv.Run(context.Background()))
}

func TestHttpServer_RunDrainsInFlightRequests(t *testing.T) {
	port := freePort(t)
	srv := NewHttpServer(config.Config{
		Env:    "development",
		Server: config.Server{Host: "127.0.0.1", Port: port},
	})

	started := make(chan struct{})
	srv.Method(http.MethodGet, "/slow", func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		runErr <- srv.Run(ctx)
	}()

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	defer client.CloseIdleConnections()
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	require.Eventually(t, func() bool {
		resp, err := client.Get(baseURL + "/health")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 2*time.Second, 20*time.Millisecond)

	requestDone := make(chan error, 1)
	go func() {
		resp, err := client.Get(baseURL + "/slow")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("unexpected status: %d", resp.StatusCode)
			}
		}
		requestDone <- err
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("slow request never reached the handler")
	}

	cancel()

	select {
	case <-srv.ShutdownSignal():
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown signal was not closed")
	}

	select {
	case err := <-requestDone:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request did not finish during drain")
	}

	select {
	case err := <-runErr:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after shutdown")
	}
}
