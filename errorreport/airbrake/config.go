package airbrake

import (
	"errors"

	"github.com/airbrake/gobrake/v5"
	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/env"
)

var (
	IMGPROXY_AIRBRAKE_PROJECT_ID  = env.Int("IMGPROXY_AIRBRAKE_PROJECT_ID")
	IMGPROXY_AIRBRAKE_PROJECT_KEY = env.String("IMGPROXY_AIRBRAKE_PROJECT_KEY")
	IMGPROXY_AIRBRAKE_ENVIRONMENT = env.String("IMGPROXY_AIRBRAKE_ENVIRONMENT")
)

// Config holds Airbrake-related configuration.
type Config struct {
	ProjectID  int
	ProjectKey string
	Env        string

	// Filter, if set, is registered on the notifier via AddFilter. Nil
	// leaves the notifier without a filter. Intended for intercepting
	// notices in tests.
	Filter func(*gobrake.Notice) *gobrake.Notice
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		ProjectID:  0,
		ProjectKey: "",
		Env:        "production",
		Filter:     nil,
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_AIRBRAKE_PROJECT_ID.Parse(&c.ProjectID),
		IMGPROXY_AIRBRAKE_PROJECT_KEY.Parse(&c.ProjectKey),
		IMGPROXY_AIRBRAKE_ENVIRONMENT.Parse(&c.Env),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for airbrake config currently
	return nil
}
