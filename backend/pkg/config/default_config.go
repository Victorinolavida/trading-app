package config

import "time"

type Database struct {
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Name     string `koanf:"name"`
	MaxConns int    `koanf:"max_conns"`
	MinConns int    `koanf:"min_conns"`
}
type GRPCServer struct {
	Protocol    string `koanf:"protocol"`
	Host        string `koanf:"host"`
	Port        int    `koanf:"port"`
	ServiceName string `koanf:"service"`
}

type GRPCClient struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type TokenConfig struct {
	Secret   string        `koanf:"secret"`
	Issuer   string        `koanf:"issuer"`
	Audience string        `koanf:"audience"`
	TTL      time.Duration `koanf:"ttl"`
}

type CookieConfig struct {
	Name     string        `koanf:"name"`
	CSRFName string        `koanf:"csrfname"`
	Path     string        `koanf:"path"`
	Domain   string        `koanf:"domain"`
	Secure   bool          `koanf:"secure"`
	SameSite string        `koanf:"samesite"`
	TTL      time.Duration `koanf:"ttl"`
	Secret   string        `koanf:"secret"`
}

type RedisConfig struct {
	Host     string `koanf:"host"`
	Port     string `koanf:"port"`
	DB       int    `koanf:"database"`
	Password string `koanf:"password"`
	Username string `koanf:"username"`
}

type RabbitMQ struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	VHost    string `koanf:"vhost"`
}

type Consumer struct {
	Exchange   string `koanf:"exchange"`
	Queue      string `koanf:"queue"`
	RoutingKey string `koanf:"routing_key"`
}

type Publisher struct {
	Exchange string `koanf:"exchange"`
}
