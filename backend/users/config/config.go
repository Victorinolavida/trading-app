package config

import (
	"github.com/victorinolavida/pkg/config"
	"time"
)

type Config struct {
	Env           string             `koanf:"env"`
	Database      config.Database    `koanf:"database"`
	TokenConfig   config.TokenConfig `koanf:"token"`
	GRPC          config.GRPCServer  `koanf:"grpc"`
	PasswordReset PasswordReset      `koanf:"password_reset"`
}

type PasswordReset struct {
	Enabled  bool          `koanf:"enabled"`
	Provider string        `koanf:"provider"`
	URL      string        `koanf:"url"`
	TTL      time.Duration `koanf:"ttl"`
	Mailpit  Mailpit       `koanf:"mailpit"`
	SMTP     SMTP          `koanf:"smtp"`
}

type SMTP struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	From     string `koanf:"from"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type Mailpit struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
	From string `koanf:"from"`
}
