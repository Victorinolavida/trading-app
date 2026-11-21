// Package mock provides an in-memory email sender for local development and tests.
package mock

import (
	"context"
	"sync"

	"github.com/victorinolavida/auth/internal/application/ports"
)

// Sender stores messages in memory instead of delivering them externally.
type Sender struct {
	mu       sync.Mutex
	messages []ports.Email
	onSend   func(ports.Email)
}

// NewSender optionally reports captured messages to a local development inbox.
func NewSender(onSend func(ports.Email)) *Sender {
	return &Sender{onSend: onSend}
}

var _ ports.EmailSender = (*Sender)(nil)

func (s *Sender) Send(ctx context.Context, email ports.Email) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	s.messages = append(s.messages, email)
	s.mu.Unlock()
	if s.onSend != nil {
		s.onSend(email)
	}
	return nil
}

// Messages returns a snapshot of the local inbox.
func (s *Sender) Messages() []ports.Email {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ports.Email(nil), s.messages...)
}
