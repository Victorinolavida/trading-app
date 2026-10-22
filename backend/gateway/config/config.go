package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/victorinolavida/pkg/config"
)

const (
	// minSecretLength is the shortest secret accepted for the csrf hmac key.
	// A short or empty key makes the csrf tokens forgeable.
	minSecretLength = 32
	// placeholderSecret is the value shipped in config.example.yaml.
	placeholderSecret = "change-me-to-a-long-random-csrf-secret"
)

type Server struct {
	Port         int      `koanf:"port"`
	Host         string   `koanf:"host"`
	AllowOrigins []string `koanf:"allow_origins"`
}

type Config struct {
	Env         string              `koanf:"env"`
	Server      Server              `koanf:"server"`
	AuthService config.GRPCClient   `koanf:"auth_service"`
	Token       config.TokenConfig  `koanf:"token"`
	Cookie      config.CookieConfig `koanf:"cookie"`
}

func (c Config) IsProduction() bool {
	return strings.EqualFold(c.Env, "production")
}

// Validate reports the configuration the gateway cannot safely run with.
func (c Config) Validate() error {
	if c.IsProduction() && len(c.Server.AllowOrigins) == 0 {
		return errors.New("server: allow_origins is required in production, the browser session is cookie based")
	}
	return c.validateCookie()
}

func (c Config) validateCookie() error {
	cookie := c.Cookie

	switch {
	case cookie.Name == "":
		return errors.New("cookie: name is required")
	case cookie.CSRFName == "":
		return errors.New("cookie: csrfname is required")
	case cookie.CSRFName == cookie.Name:
		return errors.New("cookie: csrfname must differ from name")
	case cookie.Path == "":
		return errors.New("cookie: path is required")
	case cookie.TTL <= 0:
		return errors.New("cookie: ttl must be positive")
	}

	if err := validateSameSite(cookie.SameSite); err != nil {
		return err
	}

	if len(cookie.Secret) < minSecretLength {
		return fmt.Errorf("cookie: secret must be at least %d bytes, got %d", minSecretLength, len(cookie.Secret))
	}
	if cookie.Secret == c.Token.Secret {
		return errors.New("cookie: secret must differ from token.secret")
	}

	if strings.EqualFold(cookie.SameSite, "none") && !cookie.Secure {
		return errors.New("cookie: samesite=none requires secure=true")
	}

	if c.IsProduction() {
		if !cookie.Secure {
			return errors.New("cookie: secure must be true in production")
		}
		if cookie.Secret == placeholderSecret {
			return errors.New("cookie: secret is still the example placeholder")
		}
	}

	return nil
}

func validateSameSite(value string) error {
	switch strings.ToLower(value) {
	case "lax", "strict", "none":
		return nil
	default:
		return fmt.Errorf("cookie: samesite must be one of lax, strict, none, got %q", value)
	}
}
