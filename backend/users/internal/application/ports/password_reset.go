package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidResetToken covers tokens that are unknown, expired, or already used.
var ErrInvalidResetToken = errors.New("invalid or expired password reset token")

type PasswordResetRepository interface {
	// Store replaces any previous reset token for the user. The digest must be
	// a cryptographic hash of a randomly generated token, never the raw token.
	Store(ctx context.Context, userID uuid.UUID, digest []byte, expiresAt time.Time) error
	// Redeem atomically consumes a valid token and changes the password hash.
	Redeem(ctx context.Context, digest, passwordHash []byte) error
}
