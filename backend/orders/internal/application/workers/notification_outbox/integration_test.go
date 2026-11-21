package notificationoutbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/victorinolavida/pkg/broker/rabbitmq"
)

func TestIntegrationPublishAndDeadLetter(t *testing.T) {
	if os.Getenv("NOTIFICATION_INTEGRATION") != "1" {
		t.Skip("set NOTIFICATION_INTEGRATION=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	schema := "notification_test_" + suffix
	if _, err := base.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = base.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migration, err := os.ReadFile("../../../adapters/out/postgres/migrations/20261121000000_notification_outbox.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, strings.Split(string(migration), "-- migrate:down")[0]); err != nil {
		t.Fatal(err)
	}
	marketMigration, err := os.ReadFile("../../../adapters/out/postgres/migrations/20261121000001_market_depth_outbox.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, strings.Split(string(marketMigration), "-- migrate:down")[0]); err != nil {
		t.Fatal(err)
	}
	client, err := rabbitmq.NewClient("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	exchange, queue := "notification.test."+suffix, "notification.test."+suffix
	cleanup, err := client.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup.Close()
	defer func() {
		_, _ = cleanup.QueueDelete(queue, false, false, false)
		_, _ = cleanup.QueueDelete(queue+".dlq", false, false, false)
		_ = cleanup.ExchangeDelete(exchange, false, false)
	}()
	w := New(pool, client, exchange)
	w.queue = queue
	id, user := uuid.New(), uuid.New()
	key := fmt.Sprintf("notifications.private.%s.order.created", user)
	_, err = pool.Exec(ctx, `INSERT INTO notification_outbox (id,user_id,sequence,routing_key,occurred_at,payload) VALUES ($1,$2,1,$3,now(),$4)`, id, user, key, []byte(`{"schema_version":1,"event_type":"order.created"}`))
	if err != nil {
		t.Fatal(err)
	}
	if err := w.drain(ctx); err != nil {
		t.Fatal(err)
	}
	var published bool
	if err := pool.QueryRow(ctx, `SELECT published_at IS NOT NULL FROM notification_outbox WHERE id=$1`, id).Scan(&published); err != nil || !published {
		t.Fatalf("published=%v err=%v", published, err)
	}
	delivery, ok, err := cleanup.Get(queue, false)
	if err != nil || !ok {
		t.Fatalf("message missing: %v", err)
	}
	var envelope struct {
		ID       uuid.UUID `json:"id"`
		Type     string    `json:"type"`
		Sequence int64     `json:"sequence"`
	}
	if err := json.Unmarshal(delivery.Body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.ID != id || envelope.Type != key || envelope.Sequence != 1 {
		t.Fatalf("unexpected envelope: %+v", envelope)
	}
	if err := delivery.Reject(false); err != nil {
		t.Fatal(err)
	}
	for {
		dead, ok, err := cleanup.Get(queue+".dlq", true)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			if dead.MessageId != id.String() {
				t.Fatalf("unexpected dead letter %s", dead.MessageId)
			}
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("dead letter not received")
		case <-time.After(50 * time.Millisecond):
		}
	}
	marketID := uuid.New()
	marketKey := "market.depth." + uuid.NewString()
	if _, err := cleanup.QueueDeclare(queue+".market", false, true, true, false, nil); err != nil {
		t.Fatal(err)
	}
	if err := cleanup.QueueBind(queue+".market", "market.depth.*", exchange, false, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO market_depth_outbox (id,sequence,routing_key,occurred_at,payload) VALUES ($1,7,$2,now(),$3)`, marketID, marketKey, []byte(`{"security_id":"test","ask":[],"bid":[]}`)); err != nil {
		t.Fatal(err)
	}
	if err := w.drain(ctx); err != nil {
		t.Fatal(err)
	}
	market, ok, err := cleanup.Get(queue+".market", true)
	if err != nil || !ok {
		t.Fatalf("market message missing: %v", err)
	}
	if err := json.Unmarshal(market.Body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.ID != marketID || envelope.Type != marketKey || envelope.Sequence != 7 {
		t.Fatalf("unexpected market envelope: %+v", envelope)
	}
	if _, err := cleanup.QueueDelete(queue+".market", false, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO market_depth_outbox (id,sequence,routing_key,occurred_at,payload) VALUES ($1,8,$2,now(),'{}')`, uuid.New(), marketKey); err != nil {
		t.Fatal(err)
	}
	if err := w.drain(ctx); err != nil {
		t.Fatalf("market event without subscribers: %v", err)
	}
}
