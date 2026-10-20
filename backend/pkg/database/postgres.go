package database

import (
	"context"
	"fmt"
	"github.com/victorinolavida/pkg/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func buildDSN(user, password, host string, port int, database string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, database)
}

const (
	defaultMaxConns = 20
	defaultMinConns = 2
)

func NewPool(ctx context.Context, conf config.Database) (*pgxpool.Pool, error) {
	dsn := buildDSN(conf.User, conf.Password, conf.Host, conf.Port, conf.Name)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = defaultMaxConns
	cfg.MinConns = defaultMinConns
	if conf.MaxConns > 0 {
		cfg.MaxConns = int32(conf.MaxConns)
	}
	if conf.MinConns > 0 {
		cfg.MinConns = int32(conf.MinConns)
	}

	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnLifetimeJitter = 5 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
