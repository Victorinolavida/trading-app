package config

import (
	"github.com/victorinolavida/pkg/config"
	"time"
)

type Config struct {
	Database    config.Database   `koanf:"database"`
	TokenConfig TokenConfig       `koanf:"token"`
	Grcp        config.GRPCServer `koanf:"grcp"`
}

type TokenConfig struct {
	Secret   string        `koanf:"secret"`
	Issuer   string        `koanf:"issuer"`
	Audience string        `koanf:"audience"`
	TTL      time.Duration `koanf:"ttl"`
}
