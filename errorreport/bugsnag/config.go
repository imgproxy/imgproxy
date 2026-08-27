package bugsnag

import (
	"errors"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/env"
)

var (
	IMGPROXY_BUGSNAG_KEY   = env.String("IMGPROXY_BUGSNAG_KEY")
	IMGPROXY_BUGSNAG_STAGE = env.String("IMGPROXY_BUGSNAG_STAGE")
)

// Config holds Bugsnag-related configuration.
type Config struct {
	Key   string
	Stage string

	// Endpoints overrides the notifier's API endpoints. Zero value uses
	// Bugsnag's default production endpoints. Intended for pointing tests
	// at a local mock server.
	Endpoints bugsnag.Endpoints
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		Key:       "",
		Stage:     "production",
		Endpoints: bugsnag.Endpoints{},
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_BUGSNAG_KEY.Parse(&c.Key),
		IMGPROXY_BUGSNAG_STAGE.Parse(&c.Stage),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for bugsnag config currently
	return nil
}
