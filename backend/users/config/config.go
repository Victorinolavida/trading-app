package config

import "pkg/config"

type Config struct {
	Database config.Database `koanf:"database"`
}
