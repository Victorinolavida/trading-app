package ports

import "context"

// Email is a message ready for delivery by an email provider.
type Email struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// EmailSender delivers email independently of the provider implementation.
type EmailSender interface {
	Send(ctx context.Context, email Email) error
}
