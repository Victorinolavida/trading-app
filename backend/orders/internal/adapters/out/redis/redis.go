package redis

import (
	"context"

	"github.com/victorinolavida/orders/internal/application/ports"
	"github.com/victorinolavida/pkg/config"
	"github.com/victorinolavida/pkg/redis"
)

type RedisClient struct {
	c *redis.Client
}

func New(ctx context.Context, conf config.RedisConfig) (ports.MemoryPort, error) {
	c, err := redis.NewClient(ctx, conf)
	if err != nil {
		return &RedisClient{c: nil}, err
	}
	return &RedisClient{c}, nil
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.c.Rdb.Incr(ctx, key).Result()
}
