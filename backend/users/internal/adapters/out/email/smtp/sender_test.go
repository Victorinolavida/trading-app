package smtp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/auth/internal/application/ports"
)

func TestSenderDeliversMessage(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	received := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			received <- "accept: " + err.Error()
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		fmt.Fprint(conn, "220 localhost SMTP\r\n")
		var data strings.Builder
		inData := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				received <- data.String()
				return
			}
			if inData {
				if line == ".\r\n" {
					inData = false
					fmt.Fprint(conn, "250 queued\r\n")
					continue
				}
				data.WriteString(line)
				continue
			}
			switch {
			case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"), strings.HasPrefix(line, "MAIL FROM:"), strings.HasPrefix(line, "RCPT TO:"):
				fmt.Fprint(conn, "250 OK\r\n")
			case strings.HasPrefix(line, "DATA"):
				inData = true
				fmt.Fprint(conn, "354 send data\r\n")
			case strings.HasPrefix(line, "QUIT"):
				fmt.Fprint(conn, "221 bye\r\n")
				received <- data.String()
				return
			default:
				fmt.Fprint(conn, "500 unknown\r\n")
			}
		}
	}()
	host, portString, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)
	sender, err := NewSender(host, port, "noreply@example.com")
	require.NoError(t, err)
	require.NoError(t, sender.Send(context.Background(), ports.Email{To: "user@example.com", Subject: "Reset your password", Text: "http://localhost/reset?token=abc", HTML: "<p>Reset password</p>"}))
	message := <-received
	require.Contains(t, message, "To: user@example.com\r\n")
	require.Contains(t, message, "Subject: Reset your password\r\n")
	parsed, err := mail.ReadMessage(strings.NewReader(message))
	require.NoError(t, err)
	mediaType, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	require.NoError(t, err)
	require.Equal(t, "multipart/alternative", mediaType)
	parts := multipart.NewReader(parsed.Body, params["boundary"])
	plain, err := parts.NextPart()
	require.NoError(t, err)
	plainBody, err := io.ReadAll(plain)
	require.NoError(t, err)
	require.Contains(t, string(plainBody), "http://localhost/reset?token=abc")
	html, err := parts.NextPart()
	require.NoError(t, err)
	htmlBody, err := io.ReadAll(html)
	require.NoError(t, err)
	require.Contains(t, string(htmlBody), "<p>Reset password</p>")
}

func TestSenderRejectsInvalidConfigurationAndHeaders(t *testing.T) {
	_, err := NewSender("", 1025, "noreply@example.com")
	require.Error(t, err)
	_, err = NewSender("localhost", 1025, "invalid\r\nBcc: victim@example.com")
	require.Error(t, err)
	sender, err := NewSender("localhost", 1025, "noreply@example.com")
	require.NoError(t, err)
	require.Error(t, sender.Send(context.Background(), ports.Email{To: "user@example.com", Subject: "hi\r\nBcc: victim@example.com"}))
	_, err = NewAuthenticatedSender("localhost", 587, "noreply@example.com", "", "password")
	require.Error(t, err)
}

func TestAuthenticatedSenderRequiresSTARTTLS(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	commands := make(chan string, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			commands <- ""
			return
		}
		defer conn.Close()
		fmt.Fprint(conn, "220 localhost SMTP\r\n")
		reader := bufio.NewReader(conn)
		line, _ := reader.ReadString('\n')
		fmt.Fprint(conn, "250 localhost\r\n")
		more, _ := io.ReadAll(reader)
		commands <- line + string(more)
	}()
	host, portString, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	port, err := strconv.Atoi(portString)
	require.NoError(t, err)
	sender, err := NewAuthenticatedSender(host, port, "noreply@example.com", "user", "secret")
	require.NoError(t, err)
	err = sender.Send(context.Background(), ports.Email{To: "user@example.com", Text: "hello"})
	require.ErrorContains(t, err, "does not support STARTTLS")
	require.NotContains(t, <-commands, "AUTH")
}
