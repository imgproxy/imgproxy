package prometheus

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_PROMETHEUS_BIND      = env.String("IMGPROXY_PROMETHEUS_BIND")
	IMGPROXY_PROMETHEUS_NAMESPACE = env.String("IMGPROXY_PROMETHEUS_NAMESPACE")
)

// Config holds the configuration for Prometheus monitoring
type Config struct {
	Bind      string // Prometheus server bind address
	Namespace string // Prometheus metrics namespace
}

// NewDefaultConfig returns a new default configuration for Prometheus monitoring
func NewDefaultConfig() Config {
	return Config{
		Bind:      "",
		Namespace: "",
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_PROMETHEUS_BIND.Parse(&c.Bind),
		IMGPROXY_PROMETHEUS_NAMESPACE.Parse(&c.Namespace),
	)

	return c, err
}

// Enabled returns true if Prometheus monitoring is enabled
func (c *Config) Enabled() bool {
	return len(c.Bind) > 0
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if !c.Enabled() {
		return nil
	}

	return nil
}
