// Package smtp delivers mail to Mailpit or an authenticated SMTP provider.
package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/victorinolavida/auth/internal/application/ports"
)

type Sender struct {
	host, address, from string
	username, password  string
	secure              bool
}

// NewAuthenticatedSender requires STARTTLS and SMTP authentication.
func NewAuthenticatedSender(host string, port int, from, username, password string) (*Sender, error) {
	s, err := NewSender(host, port, from)
	if err != nil {
		return nil, err
	}
	if username == "" || password == "" {
		return nil, fmt.Errorf("SMTP username and password are required")
	}
	s.username, s.password, s.secure = username, password, true
	return s, nil
}

func NewSender(host string, port int, from string) (*Sender, error) {
	if strings.TrimSpace(host) == "" || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid SMTP host or port")
	}
	address, err := mail.ParseAddress(from)
	if err != nil || address.Address != from || strings.ContainsAny(from, "\r\n") {
		return nil, fmt.Errorf("invalid SMTP sender address")
	}
	return &Sender{host: host, address: net.JoinHostPort(host, strconv.Itoa(port)), from: from}, nil
}

func (s *Sender) Send(ctx context.Context, message ports.Email) error {
	to, err := mail.ParseAddress(message.To)
	if err != nil || to.Address != message.To || strings.ContainsAny(message.To, "\r\n") || strings.ContainsAny(message.Subject, "\r\n") {
		return fmt.Errorf("invalid email recipient or subject")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", s.address)
	if err != nil {
		return err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	defer client.Close()
	if s.secure {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
		if err = client.StartTLS(&tls.Config{ServerName: s.host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("SMTP server does not support authentication")
		}
		if err = client.Auth(smtp.PlainAuth("", s.username, s.password, s.host)); err != nil {
			return err
		}
	}
	if err = client.Mail(s.from); err != nil {
		return err
	}
	if err = client.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	var body bytes.Buffer
	fmt.Fprintf(&body, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", s.from, to.Address, mime.QEncoding.Encode("utf-8", message.Subject))
	multi := multipart.NewWriter(&body)
	fmt.Fprintf(&body, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", multi.Boundary())
	for _, part := range []struct{ contentType, content string }{{"text/plain", message.Text}, {"text/html", message.HTML}} {
		if part.content == "" {
			continue
		}
		header := textproto.MIMEHeader{"Content-Type": {part.contentType + "; charset=UTF-8"}, "Content-Transfer-Encoding": {"quoted-printable"}}
		pw, err := multi.CreatePart(header)
		if err != nil {
			return err
		}
		encoded := quotedprintable.NewWriter(pw)
		if _, err = encoded.Write([]byte(part.content)); err != nil {
			return err
		}
		if err = encoded.Close(); err != nil {
			return err
		}
	}
	if err = multi.Close(); err != nil {
		return err
	}
	if _, err = w.Write(body.Bytes()); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
