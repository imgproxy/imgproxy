package newrelic

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_NEW_RELIC_APP_NAME           = env.String("IMGPROXY_NEW_RELIC_APP_NAME")
	IMGPROXY_NEW_RELIC_KEY                = env.String("IMGPROXY_NEW_RELIC_KEY")
	IMGPROXY_NEW_RELIC_LABELS             = env.StringMap("IMGPROXY_NEW_RELIC_LABELS")
	IMGPROXY_NEW_RELIC_PROPAGATE_EXTERNAL = env.Bool("IMGPROXY_NEW_RELIC_PROPAGATE_EXTERNAL")
)

// Config holds the configuration for New Relic monitoring
type Config struct {
	AppName         string            // New Relic application name
	Key             string            // New Relic license key (non-empty value enables New Relic)
	Labels          map[string]string // New Relic labels/tags
	PropagateExt    bool              // Enable propagation of tracing headers for external services
	MetricsInterval time.Duration     // Interval for sending metrics to New Relic
}

// NewDefaultConfig returns a new default configuration for New Relic monitoring
func NewDefaultConfig() Config {
	return Config{
		AppName:         "imgproxy",
		Key:             "",
		Labels:          make(map[string]string),
		PropagateExt:    false,
		MetricsInterval: 10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_NEW_RELIC_APP_NAME.Parse(&c.AppName),
		IMGPROXY_NEW_RELIC_KEY.Parse(&c.Key),
		IMGPROXY_NEW_RELIC_LABELS.Parse(&c.Labels),
		IMGPROXY_NEW_RELIC_PROPAGATE_EXTERNAL.Parse(&c.PropagateExt),
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
