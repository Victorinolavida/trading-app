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
