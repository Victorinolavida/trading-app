package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/victorinolavida/proto/notification"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"module/internal/adapters/out/postgres"
)

type Repository interface {
	List(context.Context, uuid.UUID, bool, *postgres.Cursor, int) ([]postgres.Notification, error)
	UnreadCount(context.Context, uuid.UUID) (int64, error)
	MarkRead(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	MarkAllRead(context.Context, uuid.UUID) (int64, error)
}

type Handler struct {
	pb.UnimplementedNotificationServiceServer
	repo Repository
}

func New(repo Repository) *Handler { return &Handler{repo: repo} }

func userID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return id, nil
}

func (h *Handler) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	user, err := userID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = 50
	}
	if limit < 1 || limit > 100 {
		return nil, status.Error(codes.InvalidArgument, "limit must be between 1 and 100")
	}
	var cursor *postgres.Cursor
	if c := req.GetCursor(); c != nil {
		id, err := uuid.Parse(c.GetId())
		if err != nil || id == uuid.Nil || c.GetCreatedAt() == nil || c.GetCreatedAt().CheckValid() != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid cursor")
		}
		cursor = &postgres.Cursor{ID: id, CreatedAt: c.GetCreatedAt().AsTime()}
	}
	items, err := h.repo.List(ctx, user, req.GetUnreadOnly(), cursor, limit+1)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list notifications")
	}
	more := len(items) > limit
	if more {
		items = items[:limit]
	}
	response := &pb.ListResponse{Notifications: make([]*pb.Notification, 0, len(items))}
	for _, n := range items {
		item := &pb.Notification{Id: n.ID.String(), EventType: n.EventType, PayloadJson: string(n.Payload), CreatedAt: timestamppb.New(n.CreatedAt)}
		if n.ReadAt != nil {
			item.ReadAt = timestamppb.New(*n.ReadAt)
		}
		response.Notifications = append(response.Notifications, item)
	}
	if more {
		last := items[len(items)-1]
		response.NextCursor = &pb.Cursor{Id: last.ID.String(), CreatedAt: timestamppb.New(last.CreatedAt)}
	}
	return response, nil
}

func (h *Handler) UnreadCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	user, err := userID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	count, err := h.repo.UnreadCount(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to count notifications")
	}
	return &pb.CountResponse{Count: count}, nil
}

func (h *Handler) MarkRead(ctx context.Context, req *pb.MarkReadRequest) (*pb.MarkReadResponse, error) {
	user, err := userID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(req.GetNotificationId())
	if err != nil || id == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification id")
	}
	ok, err := h.repo.MarkRead(ctx, user, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark notification read")
	}
	if !ok {
		return nil, status.Error(codes.NotFound, "notification not found")
	}
	return &pb.MarkReadResponse{}, nil
}

func (h *Handler) MarkAllRead(ctx context.Context, req *pb.UserRequest) (*pb.MarkAllReadResponse, error) {
	user, err := userID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	count, err := h.repo.MarkAllRead(ctx, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to mark notifications read")
	}
	return &pb.MarkAllReadResponse{Updated: count}, nil
}
