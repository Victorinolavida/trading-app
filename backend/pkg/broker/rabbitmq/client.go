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

const reconnectDelay = 5 * time.Second

var (
	errNotConnected  = errors.New("rabbitmq: not connected")
	errAlreadyClosed = errors.New("rabbitmq: already closed")
)

// Client manages the AMQP connection and handles automatic reconnection.
// Publisher and Consumer each open their own channel via Channel().
type Client struct {
	m               sync.Mutex
	addr            string
	connection      *amqp.Connection
	done            chan struct{}
	notifyConnClose chan *amqp.Error
	isReady         bool
}

func NewClientFromConfig(cfg config.RabbitMQ) *Client {
	vhost := cfg.VHost
	if vhost == "" {
		vhost = "/"
	}
	addr := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, vhost)
	return NewClient(addr)
}

func NewClient(addr string) *Client {
	c := &Client{
		addr: addr,
		done: make(chan struct{}),
	}
	go c.handleReconnect()
	return c
}

func (c *Client) handleReconnect() {
	for {
		c.m.Lock()
		c.isReady = false
		c.m.Unlock()

		slog.Info("rabbitmq: attempting to connect")
		if err := c.connect(); err != nil {
			slog.Error("rabbitmq: connect failed", slog.Any("error", err))
			select {
			case <-c.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		select {
		case <-c.done:
			return
		case <-c.notifyConnClose:
			slog.Warn("rabbitmq: connection closed, reconnecting...")
		}
	}
}

func (c *Client) connect() error {
	conn, err := amqp.Dial(c.addr)
	if err != nil {
		return err
	}
	c.m.Lock()
	c.connection = conn
	c.notifyConnClose = make(chan *amqp.Error, 1)
	c.connection.NotifyClose(c.notifyConnClose)
	c.isReady = true
	c.m.Unlock()
	slog.Info("rabbitmq: connected")
	return nil
}

// Channel returns a new AMQP channel. Publisher and Consumer call this
// during their own setup to get an independent channel each.
func (c *Client) Channel() (*amqp.Channel, error) {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.isReady {
		return nil, errNotConnected
	}
	return c.connection.Channel()
}

func (c *Client) IsReady() bool {
	c.m.Lock()
	defer c.m.Unlock()
	return c.isReady
}

// Close shuts down the connection cleanly.
func (c *Client) Close() error {
	c.m.Lock()
	defer c.m.Unlock()
	if !c.isReady {
		return errAlreadyClosed
	}
	close(c.done)
	err := c.connection.Close()
	c.isReady = false
	return err
}
