package config

import (
	"github.com/victorinolavida/pkg/config"
)

type Config struct {
	Database    config.Database    `koanf:"database"`
	TokenConfig config.TokenConfig `koanf:"token"`
	GRPC        config.GRPCServer  `koanf:"grpc"`
}
