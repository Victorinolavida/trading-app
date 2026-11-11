package redis

import (
	"context"
	"net"

	redisLib "github.com/redis/go-redis/v9"
	"github.com/victorinolavida/pkg/config"
)

func parseAddr(host, port string) string {
	return net.JoinHostPort(host, port)
}

type Client struct {
	Rdb *redisLib.Client
}

func NewClient(ctx context.Context, conf config.RedisConfig) (*Client, error) {
	addr := parseAddr(conf.Host, conf.Port)
	rdb := redisLib.NewClient(&redisLib.Options{
		Addr:     addr,
		Password: conf.Password, // no password set
		DB:       conf.DB,       // use default DB
		Username: conf.Username,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb}, nil
}

func (c *Client) Close() error {
	return c.Rdb.Close()
}
