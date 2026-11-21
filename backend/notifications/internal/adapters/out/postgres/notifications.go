package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	EventID   uuid.UUID
	EventType string
	Payload   json.RawMessage
	CreatedAt time.Time
	ReadAt    *time.Time
}

type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Save is idempotent for a given event and recipient. It returns whether a new
// notification was inserted, allowing callers to avoid duplicate live signals.
func (r *Repository) Save(ctx context.Context, n Notification) (bool, error) {
	if n.ID == uuid.Nil || n.UserID == uuid.Nil || n.EventID == uuid.Nil || n.EventType == "" || n.CreatedAt.IsZero() || !json.Valid(n.Payload) {
		return false, fmt.Errorf("invalid notification")
	}
	result, err := r.pool.Exec(ctx, `INSERT INTO notifications (id,user_id,event_id,event_type,payload,created_at) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (event_id,user_id) DO NOTHING`, n.ID, n.UserID, n.EventID, n.EventType, []byte(n.Payload), n.CreatedAt)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 1, nil
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, unread bool, cursor *Cursor, limit int) ([]Notification, error) {
	if userID == uuid.Nil || limit < 1 || limit > 101 {
		return nil, fmt.Errorf("invalid list arguments")
	}
	var created *time.Time
	var id *uuid.UUID
	if cursor != nil {
		if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() {
			return nil, fmt.Errorf("invalid cursor")
		}
		created = &cursor.CreatedAt
		id = &cursor.ID
	}
	rows, err := r.pool.Query(ctx, `SELECT id,user_id,event_id,event_type,payload,created_at,read_at FROM notifications WHERE user_id=$1 AND (NOT $2 OR read_at IS NULL) AND ($3::timestamptz IS NULL OR (created_at,id) < ($3,$4::uuid)) ORDER BY created_at DESC,id DESC LIMIT $5`, userID, unread, created, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Notification, 0)
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.EventID, &n.EventType, &n.Payload, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r *Repository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	if userID == uuid.Nil {
		return 0, fmt.Errorf("user is required")
	}
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id=$1 AND read_at IS NULL`, userID).Scan(&count)
	return count, err
}

// MarkRead returns false if the notification does not belong to the user.
func (r *Repository) MarkRead(ctx context.Context, userID, id uuid.UUID) (bool, error) {
	if userID == uuid.Nil || id == uuid.Nil {
		return false, fmt.Errorf("user and notification are required")
	}
	result, err := r.pool.Exec(ctx, `UPDATE notifications SET read_at=COALESCE(read_at,now()) WHERE user_id=$1 AND id=$2`, userID, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() != 0, nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, fmt.Errorf("user is required")
	}
	result, err := r.pool.Exec(ctx, `UPDATE notifications SET read_at=now() WHERE user_id=$1 AND read_at IS NULL`, userID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
