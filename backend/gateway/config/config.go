package config

import (
	"strings"

	"github.com/victorinolavida/pkg/config"
)

type Server struct {
	Port         int      `koanf:"port"`
	Host         string   `koanf:"host"`
	AllowOrigins []string `koanf:"allow_origins"`
}

type Config struct {
	Env         string             `koanf:"env"`
	Server      Server             `koanf:"server"`
	AuthService config.GRPCClient  `koanf:"auth_service"`
	Token       config.TokenConfig `koanf:"token"`
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.Env, "production")
}
