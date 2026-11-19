// Package config holds the configuration of the modular monolith.
package config

import (
	gatewayconfig "github.com/victorinolavida/gateway/config"
	pkgconfig "github.com/victorinolavida/pkg/config"
)

// Database holds the shared connection settings. The monolith connects once to
// the shared database and gives each bounded context its own schema.
type Database struct {
	User       string `koanf:"user"`
	Password   string `koanf:"password"`
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	Name       string `koanf:"name"`
	Users      string `koanf:"users"`
	Securities string `koanf:"securities"`
	Orders     string `koanf:"orders"`
	MaxConns   int    `koanf:"max_conns"`
	MinConns   int    `koanf:"min_conns"`
}

// For builds the pgx configuration for one bounded context schema.
func (d Database) For(schema string) pkgconfig.Database {
	return pkgconfig.Database{
		User:     d.User,
		Password: d.Password,
		Host:     d.Host,
		Port:     d.Port,
		Name:     d.Name,
		Schema:   schema,
		MaxConns: d.MaxConns,
		MinConns: d.MinConns,
	}
}

// Config is the root configuration of the monolith.
type Config struct {
	Env                   string                 `koanf:"env"`
	Server                gatewayconfig.Server   `koanf:"server"`
	Database              Database               `koanf:"database"`
	Token                 pkgconfig.TokenConfig  `koanf:"token"`
	Cookie                pkgconfig.CookieConfig `koanf:"cookie"`
	Redis                 pkgconfig.RedisConfig  `koanf:"redis"`
	RabbitMQ              pkgconfig.RabbitMQ     `koanf:"rabbitmq"`
	Publisher             pkgconfig.Publisher    `koanf:"publisher"`
	AllowInternalCrossing bool                   `koanf:"allow_internal_crossing"`
}

// Gateway builds the gateway configuration from the monolith configuration.
func (c Config) Gateway() gatewayconfig.Config {
	return gatewayconfig.Config{
		Env:       c.Env,
		Server:    c.Server,
		Token:     c.Token,
		Cookie:    c.Cookie,
		RabbitMQ:  c.RabbitMQ,
		Publisher: c.Publisher,
	}
}

// Validate reports invalid configuration that would prevent the monolith from
// starting safely.
func (c Config) Validate() error {
	return c.Gateway().Validate()
}
