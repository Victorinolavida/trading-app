package ordershandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
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

func TestInitialOrderBookUpdate(t *testing.T) {
	const securityID = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	update := initialOrderBookUpdate(securityID, dtos.MarketDepthResponse{
		Bids: []*dtos.MarketDepthLevel{
			{Price: "100.5", NominalValue: "15000", Quantity: 150, OrderCounts: 2, Direction: "bid"},
		},
		Asks: []*dtos.MarketDepthLevel{
			{Price: "101.5", NominalValue: "15150", Quantity: 151, OrderCounts: 1, Direction: "ask"},
		},
	})

	require.Equal(t, "market.depth", update.Type)
	require.Equal(t, securityID, update.SecurityID)
	require.Equal(t, int64(0), update.Sequence)
	require.Len(t, update.Payload.Bids, 1)
	require.Len(t, update.Payload.Asks, 1)
	require.Equal(t, "100.5", update.Payload.Bids[0].Price)
	require.Equal(t, "15000", update.Payload.Bids[0].NominalValue)
	require.Equal(t, int64(150), update.Payload.Bids[0].Quantity)
	require.Equal(t, int64(2), update.Payload.Bids[0].OrderCounts)
	require.Equal(t, "bid", update.Payload.Bids[0].Direction)
	require.Equal(t, "101.5", update.Payload.Asks[0].Price)
	require.Equal(t, "15150", update.Payload.Asks[0].NominalValue)
}

func TestInitialOrderBookUpdate_EmptyDepth(t *testing.T) {
	update := initialOrderBookUpdate("6ba7b810-9dad-11d1-80b4-00c04fd430c8", dtos.MarketDepthResponse{})

	require.Equal(t, "market.depth", update.Type)
	require.Equal(t, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", update.SecurityID)
	require.Equal(t, int64(0), update.Sequence)
	require.Empty(t, update.Payload.Bids)
	require.Empty(t, update.Payload.Asks)
}

func TestInitialOrderBookUpdate_MarshalsToSSEFrame(t *testing.T) {
	update := initialOrderBookUpdate("6ba7b810-9dad-11d1-80b4-00c04fd430c8", dtos.MarketDepthResponse{
		Bids: []*dtos.MarketDepthLevel{
			{Price: "100.5", NominalValue: "15000", Quantity: 150, OrderCounts: 2, Direction: "bid"},
		},
	})

	data, err := json.Marshal(update)
	require.NoError(t, err)
	require.JSONEq(t, `{"type":"market.depth","security_id":"6ba7b810-9dad-11d1-80b4-00c04fd430c8","sequence":0,"payload":{"bids":[{"price":"100.5","nominal_value":"15000","quantity":150,"order_counts":2,"direction":"bid"}],"asks":[]}}`, string(data))
}
