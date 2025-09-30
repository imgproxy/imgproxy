package bugsnag

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_BUGSNAG_KEY   = env.Describe("IMGPROXY_BUGSNAG_KEY", "string")
	IMGPROXY_BUGSNAG_STAGE = env.Describe("IMGPROXY_BUGSNAG_STAGE", "string")
)

// Config holds Bugsnag-related configuration.
type Config struct {
	Key   string
	Stage string
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		Key:   "",
		Stage: "production",
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Key, IMGPROXY_BUGSNAG_KEY),
		env.String(&c.Stage, IMGPROXY_BUGSNAG_STAGE),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for bugsnag config currently
	return nil
}
