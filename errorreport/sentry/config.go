package sentry

import (
	"errors"
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/version"
)

var (
	IMGPROXY_SENTRY_DSN         = env.String("IMGPROXY_SENTRY_DSN")
	IMGPROXY_SENTRY_RELEASE     = env.String("IMGPROXY_SENTRY_RELEASE")
	IMGPROXY_SENTRY_ENVIRONMENT = env.String("IMGPROXY_SENTRY_ENVIRONMENT")
)

// Config holds Sentry-related configuration.
type Config struct {
	DSN         string
	Release     string
	Environment string
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		DSN:         "",
		Release:     fmt.Sprintf("imgproxy@%s", version.Version),
		Environment: "production",
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_SENTRY_DSN.Parse(&c.DSN),
		IMGPROXY_SENTRY_RELEASE.Parse(&c.Release),
		IMGPROXY_SENTRY_ENVIRONMENT.Parse(&c.Environment),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for sentry config currently
	return nil
}
