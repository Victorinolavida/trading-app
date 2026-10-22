package config

type Server struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type Config struct {
	Env    string `koanf:"env"`
	Server Server `koanf:"server"`
}
