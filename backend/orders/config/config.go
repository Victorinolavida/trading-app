package config

import "github.com/victorinolavida/pkg/config"

type Config struct {
	Env                   string             `koanf:"env"`
	GRPC                  config.GRPCServer  `koanf:"grpc"`
	Database              config.Database    `koanf:"database"`
	Services              Services           `koanf:"services"`
	AllowInternalCrossing bool               `koanf:"allow_internal_crossing"`
	Redis                 config.RedisConfig `koanf:"redis"`
	RabbitMQ              config.RabbitMQ    `koanf:"rabbitmq"`
	Publisher             config.Publisher   `koanf:"publisher"`
}

type Services struct {
	Users      config.GRPCClient `koanf:"users"`
	Securities config.GRPCClient `koanf:"securities"`
}
