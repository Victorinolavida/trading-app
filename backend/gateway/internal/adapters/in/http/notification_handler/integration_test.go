package notificationhandler

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/gateway/config"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
	cfg "github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/jwt"
)

func TestSSEPrivateDelivery(t *testing.T) {
	if os.Getenv("NOTIFICATION_INTEGRATION") != "1" {
		t.Skip("set NOTIFICATION_INTEGRATION=1 to run")
	}
	broker, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer broker.Close()
	conf := &config.Config{Token: cfg.TokenConfig{Secret: "01234567890123456789012345678901", Issuer: "test", Audience: "test", TTL: time.Hour}, Cookie: cfg.CookieConfig{Name: "session"}}
	h, err := New(conf, &fakeClient{}, broker)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(h.Routes())
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connect := func(user string) <-chan string {
		token, err := h.token.Issue(ctx, jwt.Claims{UserID: user})
		if err != nil {
			t.Fatal(err)
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/stream", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("stream status %d", res.StatusCode)
		}
		lines := make(chan string, 2)
		go func() {
			defer res.Body.Close()
			scanner := bufio.NewScanner(res.Body)
			for scanner.Scan() {
				select {
				case lines <- scanner.Text():
				case <-ctx.Done():
					return
				}
			}
		}()
		return lines
	}
	user, other := uuid.NewString(), uuid.NewString()
	a := connect(user)
	b := connect(user)
	c := connect(other)
	ch, err := broker.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()
	if err := ch.PublishWithContext(ctx, "notifications.live", user, false, false, amqp.Publishing{Body: []byte(`{}`)}); err != nil {
		t.Fatal(err)
	}
	for _, lines := range []<-chan string{a, b} {
		select {
		case line := <-lines:
			if line != "event: notifications.changed" {
				t.Fatalf("unexpected line %q", line)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("signal not delivered")
		}
	}
	select {
	case line := <-c:
		t.Fatalf("other user received %q", line)
	case <-time.After(200 * time.Millisecond):
	}
	cancel()
}
