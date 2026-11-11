package ordershandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
)

func TestFlusherFrom_UnwrapsRequestWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := response.NewRequestWriter(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	flusher, ok := flusherFrom(w)
	require.True(t, ok, "the wrapped response writer must expose a flusher")
	require.NotNil(t, flusher)
}

func TestFlusherFrom_PlainWriter(t *testing.T) {
	flusher, ok := flusherFrom(httptest.NewRecorder())
	require.True(t, ok, "a plain response writer must expose a flusher")
	require.NotNil(t, flusher)
}
