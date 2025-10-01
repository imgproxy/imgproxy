package newrelic

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_NEW_RELIC_APP_NAME = env.Describe("IMGPROXY_NEW_RELIC_APP_NAME", "string")
	IMGPROXY_NEW_RELIC_KEY      = env.Describe("IMGPROXY_NEW_RELIC_KEY", "string")
	IMGPROXY_NEW_RELIC_LABELS   = env.Describe("IMGPROXY_NEW_RELIC_LABELS", "semicolon-separated list of key=value pairs")
)

// Config holds the configuration for New Relic monitoring
type Config struct {
	AppName         string            // New Relic application name
	Key             string            // New Relic license key (non-empty value enables New Relic)
	Labels          map[string]string // New Relic labels/tags
	MetricsInterval time.Duration     // Interval for sending metrics to New Relic
}

// NewDefaultConfig returns a new default configuration for New Relic monitoring
func NewDefaultConfig() Config {
	return Config{
		AppName:         "imgproxy",
		Key:             "",
		Labels:          make(map[string]string),
		MetricsInterval: 10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.AppName, IMGPROXY_NEW_RELIC_APP_NAME),
		env.String(&c.Key, IMGPROXY_NEW_RELIC_KEY),
		env.StringMap(&c.Labels, IMGPROXY_NEW_RELIC_LABELS),
	)

	return c, err
}

// Enabled returns true if New Relic is enabled
func (c *Config) Enabled() bool {
	return len(c.Key) > 0
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	// If Key is empty, New Relic is disabled, so no need to validate further
	if !c.Enabled() {
		return nil
	}

	// AppName should not be empty if New Relic is enabled
	if len(c.AppName) == 0 {
		return IMGPROXY_NEW_RELIC_APP_NAME.ErrorEmpty()
	}

	return nil
}
