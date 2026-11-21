package notificationhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/gateway/config"
	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/jwt"
	pb "github.com/victorinolavida/proto/notification"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeClient struct {
	user string
	id   string
}

func (f *fakeClient) List(_ context.Context, r *pb.ListRequest, _ ...grpc.CallOption) (*pb.ListResponse, error) {
	f.user = r.UserId
	return &pb.ListResponse{Notifications: []*pb.Notification{{Id: f.id, EventType: "order.created", PayloadJson: `{"order":{"id":"order"}}`, CreatedAt: timestamppb.Now()}}}, nil
}
func (f *fakeClient) UnreadCount(_ context.Context, r *pb.UserRequest, _ ...grpc.CallOption) (*pb.CountResponse, error) {
	f.user = r.UserId
	return &pb.CountResponse{Count: 1}, nil
}
func (f *fakeClient) MarkRead(_ context.Context, r *pb.MarkReadRequest, _ ...grpc.CallOption) (*pb.MarkReadResponse, error) {
	f.user = r.UserId
	f.id = r.NotificationId
	return &pb.MarkReadResponse{}, nil
}
func (f *fakeClient) MarkAllRead(_ context.Context, r *pb.UserRequest, _ ...grpc.CallOption) (*pb.MarkAllReadResponse, error) {
	f.user = r.UserId
	return &pb.MarkAllReadResponse{}, nil
}

func TestAuthenticatedRoutesUseTokenIdentity(t *testing.T) {
	conf := &config.Config{Token: cfg.TokenConfig{Secret: "01234567890123456789012345678901", Issuer: "test", Audience: "test", TTL: time.Hour}, Cookie: cfg.CookieConfig{Name: "session"}}
	client := &fakeClient{id: uuid.NewString()}
	h, err := New(conf, client, nil)
	if err != nil {
		t.Fatal(err)
	}
	user := uuid.NewString()
	token, err := h.token.(*jwt.TokenService).Issue(context.Background(), jwt.Claims{UserID: user})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/?user_id="+uuid.NewString(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	h.Routes().ServeHTTP(res, req)
	if res.Code != http.StatusOK || client.user != user {
		t.Fatalf("identity or status: %d %s", res.Code, client.user)
	}
	var body struct {
		Notifications []struct {
			Payload map[string]any `json:"payload"`
		} `json:"notifications"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil || len(body.Notifications) != 1 || body.Notifications[0].Payload["order"] == nil {
		t.Fatalf("invalid JSON response: %s %v", res.Body.String(), err)
	}
	res = httptest.NewRecorder()
	h.Routes().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/unread-count", nil))
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status %d", res.Code)
	}
}
