package errorreport

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/errorreport/airbrake"
	"github.com/imgproxy/imgproxy/v3/errorreport/bugsnag"
	"github.com/imgproxy/imgproxy/v3/errorreport/honeybadger"
	"github.com/imgproxy/imgproxy/v3/errorreport/sentry"
)

// Config holds error reporting-related configuration for all providers.
type Config struct {
	Airbrake    airbrake.Config
	Bugsnag     bugsnag.Config
	Honeybadger honeybadger.Config
	Sentry      sentry.Config
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		Airbrake:    airbrake.NewDefaultConfig(),
		Bugsnag:     bugsnag.NewDefaultConfig(),
		Honeybadger: honeybadger.NewDefaultConfig(),
		Sentry:      sentry.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	var airbErr, bugErr, honeyErr, sentErr error

	_, airbErr = airbrake.LoadConfigFromEnv(&c.Airbrake)
	_, bugErr = bugsnag.LoadConfigFromEnv(&c.Bugsnag)
	_, honeyErr = honeybadger.LoadConfigFromEnv(&c.Honeybadger)
	_, sentErr = sentry.LoadConfigFromEnv(&c.Sentry)

	err := errors.Join(airbErr, bugErr, honeyErr, sentErr)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	return errors.Join(
		c.Airbrake.Validate(),
		c.Bugsnag.Validate(),
		c.Honeybadger.Validate(),
		c.Sentry.Validate(),
	)
}
