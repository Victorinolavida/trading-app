package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestEnqueueNotificationRequiresTransaction(t *testing.T) {
	_, _, err := EnqueueNotification(context.Background(), uuid.New(), "notifications.private.user.order.created", struct{}{})
	if !errors.Is(err, ErrNoTx) {
		t.Fatalf("expected ErrNoTx, got %v", err)
	}
}
