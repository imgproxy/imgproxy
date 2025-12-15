package honeybadger

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_HONEYBADGER_KEY = env.String("IMGPROXY_HONEYBADGER_KEY")
	IMGPROXY_HONEYBADGER_ENV = env.String("IMGPROXY_HONEYBADGER_ENV")
)

// Config holds Honeybadger-related configuration.
type Config struct {
	Key string
	Env string
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		Key: "",
		Env: "production",
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_HONEYBADGER_KEY.Parse(&c.Key),
		IMGPROXY_HONEYBADGER_ENV.Parse(&c.Env),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for honeybadger config currently
	return nil
}
