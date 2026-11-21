package config

import cfg "github.com/victorinolavida/pkg/config"

type Server struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type Consumer struct {
	Exchange string `koanf:"exchange"`
	Queue    string `koanf:"queue"`
}

type Config struct {
	Env      string         `koanf:"env"`
	Server   Server         `koanf:"server"`
	GRPC     cfg.GRPCServer `koanf:"grpc"`
	Database cfg.Database   `koanf:"database"`
	RabbitMQ cfg.RabbitMQ   `koanf:"rabbitmq"`
	Consumer Consumer       `koanf:"consumer"`
}
