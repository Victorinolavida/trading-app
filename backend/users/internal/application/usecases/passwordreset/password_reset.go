package passwordreset

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/victorinolavida/auth/internal/application/ports"
	"github.com/victorinolavida/auth/internal/domain"
	"github.com/victorinolavida/auth/internal/domainerrors"
)

type Usecase struct {
	users  ports.UserRepository
	resets ports.PasswordResetRepository
	email  ports.EmailSender
	url    *url.URL
	ttl    time.Duration
}

// New requires a configured public reset-page URL, never a URL from a request.
func New(users ports.UserRepository, resets ports.PasswordResetRepository, email ports.EmailSender, resetURL string, ttl time.Duration) (*Usecase, error) {
	u, err := url.Parse(resetURL)
	if err != nil || u == nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil || u.Fragment != "" || ttl <= 0 {
		return nil, errors.New("invalid password reset configuration")
	}
	return &Usecase{users: users, resets: resets, email: email, url: u, ttl: ttl}, nil
}

// Request creates and sends a single-use reset link for a registered address.
// The caller must respond identically for registered and unregistered addresses,
// including when delivery fails, and rate-limit requests.
func (u *Usecase) Request(ctx context.Context, address string) error {
	email, err := domain.NewEmail(address)
	if err != nil {
		return err
	}
	user, err := u.users.GetByEmail(ctx, email)
	if errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(random[:])
	digest := sha256.Sum256([]byte(token))
	if err := u.resets.Store(ctx, user.ID(), digest[:], time.Now().Add(u.ttl)); err != nil {
		return err
	}
	link := *u.url
	query := link.Query()
	query.Set("token", token)
	link.RawQuery = query.Encode()
	return u.email.Send(ctx, ports.Email{
		To: address, Subject: "Reset your password",
		Text: fmt.Sprintf("Use this link to reset your password: %s", link.String()),
	})
}

func (u *Usecase) Confirm(ctx context.Context, token, newPassword string) error {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return ports.ErrInvalidResetToken
	}
	password, err := domain.NewPassword(newPassword)
	if err != nil {
		return err
	}
	digest := sha256.Sum256([]byte(token))
	return u.resets.Redeem(ctx, digest[:], password.Hash())
}
