package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/victorinolavida/orders/internal/domain"
)

type NotificationOutbox struct{}

func NewNotificationOutbox() *NotificationOutbox { return &NotificationOutbox{} }

func (*NotificationOutbox) EnqueueOrder(ctx context.Context, eventType string, order *domain.Order, reason string) error {
	if !strings.HasPrefix(eventType, "order.") || strings.Contains(strings.TrimPrefix(eventType, "order."), ".") {
		return fmt.Errorf("invalid notification event type %q", eventType)
	}
	payload := struct {
		SchemaVersion int    `json:"schema_version"`
		EventType     string `json:"event_type"`
		UserID        string `json:"user_id"`
		Reason        string `json:"reason,omitempty"`
		Order         struct {
			ID               string `json:"id"`
			SecurityID       string `json:"security_id"`
			Status           string `json:"status"`
			Price            string `json:"price"`
			Yield            string `json:"yield"`
			Quantity         int64  `json:"quantity"`
			AllocationStatus string `json:"allocation_status"`
		} `json:"order"`
	}{SchemaVersion: 1, EventType: eventType, UserID: order.TraderID().String(), Reason: reason}
	payload.Order.ID = order.PublicID().String()
	payload.Order.SecurityID = order.SecurityID().String()
	payload.Order.Status = order.Status().String()
	payload.Order.Price = domain.FormatDecimal(order.Price())
	payload.Order.Yield = domain.FormatDecimal(order.Yield())
	payload.Order.Quantity = order.Quantity().Value()
	payload.Order.AllocationStatus = order.AllocationStatus().String()
	routingKey := "notifications.private." + payload.UserID + "." + eventType
	_, _, err := EnqueueNotification(ctx, order.TraderID(), routingKey, payload)
	return err
}

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
