package ordershandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/victorinolavida/gateway/config"
	authmiddleware "github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/module"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/application/dtos"
	"github.com/victorinolavida/gateway/internal/application/ports"
	broker "github.com/victorinolavida/pkg/broker"
	brokerrabbit "github.com/victorinolavida/pkg/broker/rabbitmq"
	"github.com/victorinolavida/pkg/jwt"
)

var ErrMissingValidator = errors.New("orders handler: validator is required")

const (
	orderBookEventType    = "market.depth"
	defaultOrdersExchange = "orders.events"
)

type OrdersHandler struct {
	conf          *config.Config
	ordersService ports.OrdersService
	validate      *validator.Validate
	jwtService    *jwt.TokenService
	rabbitClient  *brokerrabbit.Client
}

func NewOrdersHandler(conf *config.Config, ordersService ports.OrdersService, validate *validator.Validate, rabbitClient *brokerrabbit.Client) (*OrdersHandler, error) {
	if validate == nil {
		return nil, ErrMissingValidator
	}
	jwtService, err := jwt.NewService(conf.Token)
	if err != nil {
		return nil, err
	}
	return &OrdersHandler{
		conf:          conf,
		ordersService: ordersService,
		validate:      validate,
		jwtService:    jwtService,
		rabbitClient:  rabbitClient,
	}, nil
}

var _ module.Module = (*OrdersHandler)(nil)

func (h *OrdersHandler) Routes() http.Handler {
	r := chi.NewMux()
	protected := r.With(authmiddleware.AuthMiddleware(h.jwtService, &h.conf.Cookie))
	protected.Post("/", h.CreateOrder)
	protected.Patch("/{orderId}/cancel", h.CancelOrder)
	protected.Get("/active", h.GetActiveOrders)
	protected.Get("/", h.ListOrders)
	protected.Get("/securities/{securityId}/book", h.StreamOrderBook)
	return r
}

func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	req, err := response.ParseBodyFromRequest[dtos.CreateOrderRequest](w, r)
	if err != nil {
		response.RenderBadRequest(w, err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	// Get trader from auth context if available
	if claims, err := authmiddleware.GetClaims(r.Context()); err == nil {
		if req.TraderID == "" {
			req.TraderID = claims.UserID
		}
	}

	res, err := h.ordersService.CreateOrder(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusCreated)
}

func (h *OrdersHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		slog.Error("error getting claims", slog.Any("error", err))
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}

	orderID := chi.URLParam(r, "orderId")
	req := dtos.CancelOrderRequest{
		TraderID: claims.UserID,
		OrderID:  orderID,
	}

	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	res, err := h.ordersService.CancelOrder(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}

	q := r.URL.Query()
	req := dtos.ListOrdersRequest{
		TraderID:  claims.UserID,
		OrderIDs:  parseOrderIDs(q["order_ids"]),
		Direction: q.Get("direction"),
		Status:    q.Get("status"),
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}

	res, err := h.ordersService.ListOrders(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

// parseOrderIDs accepts repeated ?order_ids=... parameters and
// comma-separated values, dropping empty entries.
func parseOrderIDs(values []string) []string {
	ids := make([]string, 0, len(values))
	for _, v := range values {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func (h *OrdersHandler) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	claims, err := authmiddleware.GetClaims(r.Context())
	if err != nil {
		response.RenderError(w, http.StatusUnauthorized, err.Error())
		return
	}
	req := dtos.GetActiveOrdersRequest{
		SecurityID: q.Get("security_id"),
		Direction:  q.Get("direction"),
		TraderID:   claims.UserID,
	}
	if err := h.validate.Struct(req); err != nil {
		response.RenderValidationError(w, err)
		return
	}
	res, err := h.ordersService.GetActiveOrders(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, res, http.StatusOK)
}

// StreamOrderBook subscribes to the order book events of a security from the
// orders event broker and forwards them to the client as server sent events.
// It first sends the current market depth snapshot as the initial event before
// streaming subsequent book updates.
func (h *OrdersHandler) StreamOrderBook(w http.ResponseWriter, r *http.Request) {
	securityID := chi.URLParam(r, "securityId")
	if _, err := uuid.Parse(securityID); err != nil {
		response.RenderBadRequest(w, "invalid security_id")
		return
	}

	if !h.rabbitClient.IsReady() {
		response.RenderError(w, http.StatusServiceUnavailable, "orders events broker is not ready")
		return
	}

	flusher, ok := flusherFrom(w)
	if !ok {
		response.RenderError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	ctx := r.Context()

	initialDepth, err := h.ordersService.GetMarketDepth(ctx, dtos.GetMarketDepthRequest{SecurityID: securityID})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	initialUpdate := initialOrderBookUpdate(securityID, initialDepth)
	initialData, err := json.Marshal(initialUpdate)
	if err != nil {
		slog.Error("orders: marshal initial order book update", slog.Any("error", err))
		return
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", initialData); err != nil {
		return
	}
	flusher.Flush()

	routingKey := fmt.Sprintf("%s.%s", orderBookEventType, securityID)
	exchange := h.conf.Publisher.Exchange
	if exchange == "" {
		exchange = defaultOrdersExchange
	}

	ch, err := h.rabbitClient.Channel()
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "orders events broker unavailable")
		return
	}

	queue, err := ch.QueueDeclare("", true, true, true, false, nil)
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "orders events queue error")
		return
	}
	if err := ch.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "orders events binding error")
		return
	}
	deliveries, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "orders events consume error")
		return
	}

	slog.Info("orders: streaming order book", slog.String("security_id", securityID))

	for {
		select {
		case <-ctx.Done():
			_ = ch.Close()
			return
		case d, ok := <-deliveries:
			if !ok {
				_ = ch.Close()
				return
			}

			var event broker.Event[*OrderBookEvent]
			if err := json.Unmarshal(d.Body, &event); err != nil {
				slog.Error("orders: unmarshal order book event", slog.Any("error", err))
				_ = d.Nack(false, true)
				continue
			}

			update := toOrderBookUpdate(event)
			data, err := json.Marshal(update)
			if err != nil {
				slog.Error("orders: marshal order book update", slog.Any("error", err))
				_ = d.Ack(false)
				continue
			}

			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return
			}
			if err := d.Ack(false); err != nil {
				slog.Error("orders: ack order book event", slog.Any("error", err))
			}
			flusher.Flush()
		}
	}
}

type OrderBookLevel struct {
	Price           string `json:"price"`
	AggNominalValue string `json:"nominal_value"`
	AggQuantity     int64  `json:"quantity"`
	OrderCounts     int64  `json:"order_counts"`
	Direction       string `json:"direction"`
}

type OrderBookEvent struct {
	Ask        []*OrderBookLevel `json:"ask"`
	Bid        []*OrderBookLevel `json:"bid"`
	SecurityID string            `json:"security_id"`
}

func (e *OrderBookEvent) EventType() string {
	return fmt.Sprintf("%s.%s", orderBookEventType, e.SecurityID)
}

type bookLevel struct {
	Price        string `json:"price"`
	NominalValue string `json:"nominal_value"`
	Quantity     int64  `json:"quantity"`
	OrderCounts  int64  `json:"order_counts"`
	Direction    string `json:"direction"`
}

type bookSnapshot struct {
	Bids []*bookLevel `json:"bids"`
	Asks []*bookLevel `json:"asks"`
}

type orderBookUpdate struct {
	Type       string       `json:"type"`
	SecurityID string       `json:"security_id"`
	Sequence   int64        `json:"sequence"`
	Payload    bookSnapshot `json:"payload"`
}

func toOrderBookUpdate(event broker.Event[*OrderBookEvent]) orderBookUpdate {
	update := orderBookUpdate{
		Type:     orderBookEventType,
		Sequence: event.Sequence,
	}
	if event.Payload == nil {
		return update
	}
	update.SecurityID = event.Payload.SecurityID
	for _, side := range event.Payload.Bid {
		if level := toBookLevel(side); level != nil {
			update.Payload.Bids = append(update.Payload.Bids, level)
		}
	}
	for _, side := range event.Payload.Ask {
		if level := toBookLevel(side); level != nil {
			update.Payload.Asks = append(update.Payload.Asks, level)
		}
	}
	return update
}

func toBookLevel(side *OrderBookLevel) *bookLevel {
	if side == nil {
		return nil
	}
	return &bookLevel{
		Price:        side.Price,
		NominalValue: side.AggNominalValue,
		Quantity:     side.AggQuantity,
		OrderCounts:  side.OrderCounts,
		Direction:    side.Direction,
	}
}

func initialOrderBookUpdate(securityID string, depth dtos.MarketDepthResponse) orderBookUpdate {
	update := orderBookUpdate{
		Type:       orderBookEventType,
		SecurityID: securityID,
		Sequence:   0,
		Payload: bookSnapshot{
			Bids: toBookLevels(depth.Bids),
			Asks: toBookLevels(depth.Asks),
		},
	}
	return update
}

func toBookLevels(levels []*dtos.MarketDepthLevel) []*bookLevel {
	converted := make([]*bookLevel, 0, len(levels))
	for _, level := range levels {
		if level == nil {
			continue
		}
		converted = append(converted, &bookLevel{
			Price:        level.Price,
			NominalValue: level.NominalValue,
			Quantity:     level.Quantity,
			OrderCounts:  level.OrderCounts,
			Direction:    level.Direction,
		})
	}
	return converted
}

func (h *OrdersHandler) BasePath() string {
	return "/orders"
}

// flusherFrom resolves the http.Flusher of the response writer. The gateway
// wraps every response writer in a response.RequestWriter, which does not
// implement http.Flusher itself, so the wrapped writer has to be unwrapped.
func flusherFrom(w http.ResponseWriter) (http.Flusher, bool) {
	if flusher, ok := w.(http.Flusher); ok {
		return flusher, true
	}
	if rw, ok := w.(*response.RequestWriter); ok {
		if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
			return flusher, true
		}
	}
	return nil, false
}
