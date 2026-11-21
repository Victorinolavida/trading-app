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
}
