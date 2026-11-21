package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EnqueueNotification atomically assigns a recipient sequence and saves the
// event in the transaction that owns the order change.
func EnqueueNotification(ctx context.Context, userID uuid.UUID, routingKey string, payload any) (uuid.UUID, int64, error) {
	tx, ok := txFromContext(ctx)
	if !ok {
		return uuid.Nil, 0, ErrNoTx
	}
	if userID == uuid.Nil || routingKey == "" {
		return uuid.Nil, 0, fmt.Errorf("notification recipient and routing key are required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("marshal notification: %w", err)
	}
	var sequence int64
	err = tx.QueryRow(ctx, `INSERT INTO notification_sequences (user_id, sequence)
		VALUES ($1, 1) ON CONFLICT (user_id) DO UPDATE
		SET sequence = notification_sequences.sequence + 1 RETURNING sequence`, userID).Scan(&sequence)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("allocate notification sequence: %w", err)
	}
	id := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO notification_outbox
		(id, user_id, sequence, routing_key, occurred_at, payload)
		VALUES ($1, $2, $3, $4, $5, $6)`, id, userID, sequence, routingKey, time.Now().UTC(), body)
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("save notification outbox event: %w", err)
	}
	return id, sequence, nil
}
