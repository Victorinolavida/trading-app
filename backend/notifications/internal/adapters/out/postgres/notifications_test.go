package postgres

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryIntegration(t *testing.T) {
	if os.Getenv("NOTIFICATION_INTEGRATION") != "1" {
		t.Skip("set NOTIFICATION_INTEGRATION=1 to run")
	}
	ctx := context.Background()
	url := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	base, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Close()
	schema := "notification_repo_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := base.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = base.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE") }()
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	migration, err := os.ReadFile("migrations/20261121000000_create_notifications.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, strings.Split(string(migration), "-- migrate:down")[0]); err != nil {
		t.Fatal(err)
	}
	r := New(pool)
	user, other := uuid.New(), uuid.New()
	n := Notification{ID: uuid.New(), UserID: user, EventID: uuid.New(), EventType: "order.created", Payload: []byte(`{"event_type":"order.created"}`), CreatedAt: time.Now().UTC()}
	inserted, err := r.Save(ctx, n)
	if err != nil || !inserted {
		t.Fatalf("inserted %v: %v", inserted, err)
	}
	n.ID = uuid.New()
	inserted, err = r.Save(ctx, n)
	if err != nil || inserted {
		t.Fatalf("duplicate inserted %v: %v", inserted, err)
	}
	list, err := r.List(ctx, user, true, nil, 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("list %v: %v", list, err)
	}
	if ok, err := r.MarkRead(ctx, other, list[0].ID); err != nil || ok {
		t.Fatalf("other user marked read: %v %v", ok, err)
	}
	if count, err := r.UnreadCount(ctx, user); err != nil || count != 1 {
		t.Fatalf("count %d: %v", count, err)
	}
	if ok, err := r.MarkRead(ctx, user, list[0].ID); err != nil || !ok {
		t.Fatalf("mark read %v: %v", ok, err)
	}
	list, err = r.List(ctx, user, false, nil, 10)
	if err != nil || list[0].ReadAt == nil {
		t.Fatalf("read state: %v", err)
	}
	first := *list[0].ReadAt
	if _, err = r.MarkRead(ctx, user, list[0].ID); err != nil {
		t.Fatal(err)
	}
	list, err = r.List(ctx, user, false, nil, 10)
	if err != nil || !list[0].ReadAt.Equal(first) {
		t.Fatalf("read time changed: %v", err)
	}
	if count, err := r.UnreadCount(ctx, user); err != nil || count != 0 {
		t.Fatalf("count %d: %v", count, err)
	}
}
