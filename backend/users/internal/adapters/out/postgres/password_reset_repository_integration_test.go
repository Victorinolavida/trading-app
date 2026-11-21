package postgres

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/victorinolavida/auth/internal/application/ports"
)

// Set TEST_DATABASE_URL to a disposable PostgreSQL database to run this test.
func TestPasswordResetRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer admin.Close()
	schema := "reset_test_" + uuid.NewString()[:8]
	_, err = admin.Exec(ctx, `CREATE SCHEMA `+schema)
	require.NoError(t, err)
	defer func() { _, _ = admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`) }()

	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)
	defer pool.Close()
	_, err = pool.Exec(ctx, `CREATE TABLE users (
		id uuid PRIMARY KEY, password_hash text NOT NULL CHECK (password_hash <> 'reject'),
		updated_at timestamptz NOT NULL DEFAULT now(), session_version bigint NOT NULL DEFAULT 0
	); CREATE TABLE password_reset_tokens (
		user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		token_hash bytea NOT NULL UNIQUE, expires_at timestamptz NOT NULL
	)`)
	require.NoError(t, err)
	userID := uuid.New()
	_, err = pool.Exec(ctx, `INSERT INTO users (id, password_hash) VALUES ($1, 'old')`, userID)
	require.NoError(t, err)
	repo := NewPasswordResetRepository(pool)

	require.NoError(t, repo.Store(ctx, userID, []byte("expired"), time.Now().Add(-time.Hour)))
	require.ErrorIs(t, repo.Redeem(ctx, []byte("expired"), []byte("new")), ports.ErrInvalidResetToken)
	require.NoError(t, repo.Store(ctx, userID, []byte("old-token"), time.Now().Add(time.Hour)))
	require.NoError(t, repo.Store(ctx, userID, []byte("new-token"), time.Now().Add(time.Hour)))
	require.ErrorIs(t, repo.Redeem(ctx, []byte("old-token"), []byte("new")), ports.ErrInvalidResetToken)

	// A failed password update must roll back token consumption.
	require.Error(t, repo.Redeem(ctx, []byte("new-token"), []byte("reject")))
	const attempts = 8
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- repo.Redeem(ctx, []byte("new-token"), []byte("new"))
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else {
			require.ErrorIs(t, err, ports.ErrInvalidResetToken)
		}
	}
	require.Equal(t, 1, successes)
	var hash string
	require.NoError(t, pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&hash))
	require.Equal(t, "new", hash)
	var version int64
	require.NoError(t, pool.QueryRow(ctx, `SELECT session_version FROM users WHERE id = $1`, userID).Scan(&version))
	require.EqualValues(t, 1, version)
}
