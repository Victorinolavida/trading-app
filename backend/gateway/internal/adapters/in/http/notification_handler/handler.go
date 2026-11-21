package notificationhandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/gateway/config"
	auth "github.com/victorinolavida/gateway/internal/adapters/in/http/middleware/auth_middleware"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/response"
	"github.com/victorinolavida/gateway/internal/adapters/in/http/stream"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
	"github.com/victorinolavida/pkg/jwt"
	pb "github.com/victorinolavida/proto/notification"
	"google.golang.org/protobuf/encoding/protojson"
)

type Handler struct {
	conf   *config.Config
	client pb.NotificationServiceClient
	token  *jwt.TokenService
	rabbit *rabbitmq.Client
}

func New(conf *config.Config, client pb.NotificationServiceClient, rabbit *rabbitmq.Client) (*Handler, error) {
	token, err := jwt.NewService(conf.Token)
	if err != nil {
		return nil, err
	}
	return &Handler{conf, client, token, rabbit}, nil
}
func (*Handler) BasePath() string { return "/notifications" }
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(auth.AuthMiddleware(h.token, &h.conf.Cookie))
	r.Get("/", h.list)
	r.Get("/unread-count", h.count)
	r.Post("/read-all", h.readAll)
	r.Get("/stream", h.stream)
	r.Patch("/{id}/read", h.markRead)
	return r
}
func identity(w http.ResponseWriter, r *http.Request) (string, bool) {
	claims, err := auth.GetClaims(r.Context())
	if err != nil {
		response.RenderUnauthorized(w)
		return "", false
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil || id == uuid.Nil {
		response.RenderUnauthorized(w)
		return "", false
	}
	return id.String(), true
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(w, r)
	if !ok {
		return
	}
	req := &pb.ListRequest{UserId: user}
	if v := r.URL.Query().Get("unread"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			response.RenderBadRequest(w, "invalid unread filter")
			return
		}
		req.UnreadOnly = b
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil || n < 1 || n > 100 {
			response.RenderBadRequest(w, "invalid limit")
			return
		}
		req.Limit = int32(n)
	}
	if v := r.URL.Query().Get("cursor"); v != "" {
		b, err := base64.RawURLEncoding.DecodeString(v)
		if err != nil || len(b) > 1024 {
			response.RenderBadRequest(w, "invalid cursor")
			return
		}
		req.Cursor = &pb.Cursor{}
		if err := protojson.Unmarshal(b, req.Cursor); err != nil {
			response.RenderBadRequest(w, "invalid cursor")
			return
		}
	}
	result, err := h.client.List(r.Context(), req)
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	type item struct {
		ID        string          `json:"id"`
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
		CreatedAt time.Time       `json:"created_at"`
		ReadAt    *time.Time      `json:"read_at"`
	}
	items := make([]item, 0, len(result.Notifications))
	for _, n := range result.Notifications {
		if !json.Valid([]byte(n.PayloadJson)) || n.CreatedAt == nil || n.CreatedAt.CheckValid() != nil {
			response.RenderServerError(w, fmt.Errorf("invalid stored notification"))
			return
		}
		var read *time.Time
		if n.ReadAt != nil {
			t := n.ReadAt.AsTime()
			read = &t
		}
		items = append(items, item{n.Id, n.EventType, json.RawMessage(n.PayloadJson), n.CreatedAt.AsTime(), read})
	}
	var cursor *string
	if result.NextCursor != nil {
		b, err := protojson.Marshal(result.NextCursor)
		if err != nil {
			response.RenderServerError(w, err)
			return
		}
		s := base64.RawURLEncoding.EncodeToString(b)
		cursor = &s
	}
	response.RenderJson(w, map[string]any{"notifications": items, "next_cursor": cursor}, http.StatusOK)
}
func (h *Handler) count(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(w, r)
	if !ok {
		return
	}
	result, err := h.client.UnreadCount(r.Context(), &pb.UserRequest{UserId: user})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	response.RenderJson(w, map[string]int64{"count": result.Count}, http.StatusOK)
}
func (h *Handler) readAll(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(w, r)
	if !ok {
		return
	}
	_, err := h.client.MarkAllRead(r.Context(), &pb.UserRequest{UserId: user})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(w, r)
	if !ok {
		return
	}
	_, err := h.client.MarkRead(r.Context(), &pb.MarkReadRequest{UserId: user, NotificationId: chi.URLParam(r, "id")})
	if err != nil {
		response.RenderGRPCError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	user, ok := identity(w, r)
	if !ok {
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		if rw, yes := w.(*response.RequestWriter); yes {
			flusher, ok = rw.ResponseWriter.(http.Flusher)
		}
	}
	if !ok {
		response.RenderError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	if h.rabbit == nil || !h.rabbit.IsReady() {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	ch, err := h.rabbit.Channel()
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare("notifications.live", "topic", true, false, false, false, nil); err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	q, err := ch.QueueDeclare("", false, true, true, false, amqp.Table{"x-max-length": int32(1)})
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	if err := ch.QueueBind(q.Name, user, "notifications.live", false, nil); err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		response.RenderError(w, http.StatusServiceUnavailable, "notification stream unavailable")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-stream.Done(r.Context()):
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			var signal struct {
				ID        string `json:"id"`
				EventType string `json:"event_type"`
			}
			if err := json.Unmarshal(msg.Body, &signal); err != nil {
				continue
			}
			data, err := json.Marshal(signal)
			if err != nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "event: notifications.changed\ndata: %s\n\n", data); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
