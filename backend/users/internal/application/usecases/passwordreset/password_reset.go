package passwordreset

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
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

var resetEmailHTML = template.Must(template.New("reset").Parse(`<!doctype html>
<html><body style="margin:0;background:#f3f5f8;font-family:Arial,sans-serif;color:#172033">
<div style="max-width:560px;margin:32px auto;padding:32px;background:#fff;border-radius:12px">
<p style="font-size:13px;font-weight:bold;letter-spacing:1px;color:#52637b">TRADING APP</p>
<h1 style="font-size:24px;margin:24px 0 16px">Reset your password</h1>
<p>Hi {{.Name}},</p><p>We received a request to reset your password. Select the button below to choose a new one.</p>
<p style="margin:28px 0"><a href="{{.Link}}" style="background:#244f9e;color:#fff;text-decoration:none;padding:14px 22px;border-radius:6px;display:inline-block;font-weight:bold">Reset password</a></p>
<p>This link expires in {{.TTL}} and can be used once. If you didn't request this, you can safely ignore this email.</p>
<p style="font-size:13px;color:#52637b;overflow-wrap:anywhere">If the button doesn't work, copy this link into your browser:<br>{{.Link}}</p>
</div></body></html>`))

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
	var html bytes.Buffer
	data := struct{ Name, Link, TTL string }{user.FirstName().String(), link.String(), u.ttl.String()}
	if err := resetEmailHTML.Execute(&html, data); err != nil {
		return err
	}
	return u.email.Send(ctx, ports.Email{
		To: address, Subject: "Reset your password",
		Text: fmt.Sprintf("Hi %s,\n\nWe received a request to reset your password. Use the link below to choose a new one:\n\n%s\n\nThis link expires in %s and can be used once. If you didn't request this, you can safely ignore this email.\n\nTrading App", data.Name, data.Link, data.TTL),
		HTML: html.String(),
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
