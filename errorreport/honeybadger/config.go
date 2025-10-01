package honeybadger

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_HONEYBADGER_KEY = env.Describe("IMGPROXY_HONEYBADGER_KEY", "string")
	IMGPROXY_HONEYBADGER_ENV = env.Describe("IMGPROXY_HONEYBADGER_ENV", "string")
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
		env.String(&c.Key, IMGPROXY_HONEYBADGER_KEY),
		env.String(&c.Env, IMGPROXY_HONEYBADGER_ENV),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for honeybadger config currently
	return nil
}
