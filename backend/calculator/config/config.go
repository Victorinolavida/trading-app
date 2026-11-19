package config

import "time"

type Config struct {
	Env        string           `koanf:"env"`
	Securities SecuritiesConfig `koanf:"securities"`
}

type SecuritiesConfig struct {
	BaseURL string        `koanf:"baseUrl"`
	Timeout time.Duration `koanf:"timeout"`
}
