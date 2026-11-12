package rabbitmq

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/victorinolavida/pkg/config"
)

const (
	reconnectDelay      = 5 * time.Second
	initialConnectTries = 5
	initialConnectDelay = 2 * time.Second
)

var errNotConnected = errors.New("rabbitmq: not connected")

type Client struct {
	m          sync.Mutex
	addr       string
	connection *amqp.Connection
	isReady    bool

	done chan struct{}
	once sync.Once
}

// NewClientFromConfig builds the address from config and connects.
func NewClientFromConfig(cfg config.RabbitMQ) (*Client, error) {
	vhost := cfg.VHost
	if vhost == "" {
		vhost = "/"
	}
	addr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, vhost)
	return NewClient(addr)
}

// NewClient connects synchronously (with a few retries, useful in local dev
// when the broker is still starting) and then starts the background
// reconnect loop. It returns an error if the broker cannot be reached.
func NewClient(addr string) (*Client, error) {
	c := &Client{
		addr: addr,
		done: make(chan struct{}),
	}

	var (
		notify <-chan *amqp.Error
		err    error
	)
	for attempt := 1; attempt <= initialConnectTries; attempt++ {
		notify, err = c.connect()
		if err == nil {
			go c.handleReconnect(notify)
			return c, nil
		}
		slog.Warn("rabbitmq: initial connect failed",
			slog.Int("attempt", attempt),
			slog.Any("error", err),
		)
		if attempt < initialConnectTries {
			time.Sleep(initialConnectDelay)
		}
	}
	return nil, fmt.Errorf("rabbitmq: initial connect: %w", err)
}

// connect dials the broker, stores the connection and returns the
// close-notification channel for that connection.
func (c *Client) connect() (<-chan *amqp.Error, error) {
	conn, err := amqp.Dial(c.addr)
	if err != nil {
		return nil, err
	}
	notify := conn.NotifyClose(make(chan *amqp.Error, 1))

	c.m.Lock()
	c.connection = conn
	c.isReady = true
	c.m.Unlock()

	slog.Info("rabbitmq: connected")
	return notify, nil
}

// handleReconnect waits for the current connection to drop, then retries
// until it is back or Close() is called.
func (c *Client) handleReconnect(notify <-chan *amqp.Error) {
	for {
		select {
		case <-c.done:
			return
		case err := <-notify:
			slog.Warn("rabbitmq: connection closed, reconnecting", slog.Any("error", err))
		}

		c.m.Lock()
		c.isReady = false
		c.m.Unlock()

		for {
			select {
			case <-c.done:
				return
			default:
			}

			n, err := c.connect()
			if err == nil {
				notify = n
				break
			}
			slog.Error("rabbitmq: reconnect failed", slog.Any("error", err))

			select {
			case <-c.done:
				return
			case <-time.After(reconnectDelay):
			}
		}
	}
}

// Channel returns a new AMQP channel on the shared connection.
// The mutex is not held during the network round trip.
func (c *Client) Channel() (*amqp.Channel, error) {
	c.m.Lock()
	ready, conn := c.isReady, c.connection
	c.m.Unlock()

	if !ready {
		return nil, errNotConnected
	}
	return conn.Channel()
}

// IsReady reports whether the connection is currently usable.
func (c *Client) IsReady() bool {
	c.m.Lock()
	defer c.m.Unlock()
	return c.isReady
}

// Close stops the reconnect loop and closes the connection.
// It is idempotent and safe to call while a reconnect is in progress.
func (c *Client) Close() error {
	c.once.Do(func() { close(c.done) })

	c.m.Lock()
	conn := c.connection
	c.isReady = false
	c.m.Unlock()

	if conn == nil || conn.IsClosed() {
		return nil
	}
	return conn.Close()
}
