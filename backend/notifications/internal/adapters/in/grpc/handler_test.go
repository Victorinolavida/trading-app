package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pb "github.com/victorinolavida/proto/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"module/internal/adapters/out/postgres"
)

type stubRepo struct {
	items []postgres.Notification
	user  uuid.UUID
	found bool
}

func (s *stubRepo) List(_ context.Context, u uuid.UUID, _ bool, _ *postgres.Cursor, _ int) ([]postgres.Notification, error) {
	s.user = u
	return s.items, nil
}
func (*stubRepo) UnreadCount(context.Context, uuid.UUID) (int64, error) { return 3, nil }
func (s *stubRepo) MarkRead(_ context.Context, u uuid.UUID, _ uuid.UUID) (bool, error) {
	s.user = u
	return s.found, nil
}
func (*stubRepo) MarkAllRead(context.Context, uuid.UUID) (int64, error) { return 2, nil }

func TestListPaginationAndIdentity(t *testing.T) {
	user := uuid.New()
	now := time.Now().UTC()
	repo := &stubRepo{items: []postgres.Notification{{ID: uuid.New(), CreatedAt: now, EventType: "order.created", Payload: []byte(`{}`)}, {ID: uuid.New(), CreatedAt: now, EventType: "order.created", Payload: []byte(`{}`)}}}
	res, err := New(repo).List(context.Background(), &pb.ListRequest{UserId: user.String(), Limit: 1})
	if err != nil || len(res.GetNotifications()) != 1 || res.GetNextCursor().GetId() != repo.items[0].ID.String() || repo.user != user {
		t.Fatalf("unexpected result: %v %v", res, err)
	}
	if _, err := New(repo).List(context.Background(), &pb.ListRequest{UserId: "bad"}); status.Code(err) != codes.InvalidArgument {
		t.Fatalf("invalid user: %v", err)
	}
}

func TestMarkReadDoesNotExposeOtherUsers(t *testing.T) {
	repo := &stubRepo{}
	user := uuid.New()
	_, err := New(repo).MarkRead(context.Background(), &pb.MarkReadRequest{UserId: user.String(), NotificationId: uuid.NewString()})
	if status.Code(err) != codes.NotFound || repo.user != user {
		t.Fatalf("unexpected result: %v", err)
	}
}
