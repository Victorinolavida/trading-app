package config

import "github.com/victorinolavida/pkg/config"

type Config struct {
	Env               string            `koanf:"env"`
	GRPC              config.GRPCServer `koanf:"grpc"`
	SecuritiesService config.GRPCClient `koanf:"securities_service"`
	Database          config.Database   `koanf:"database"`
}
