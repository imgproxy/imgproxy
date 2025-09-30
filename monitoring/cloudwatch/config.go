package cloudwatch

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_CLOUD_WATCH_SERVICE_NAME = env.Describe("IMGPROXY_CLOUD_WATCH_SERVICE_NAME", "string")
	IMGPROXY_CLOUD_WATCH_NAMESPACE    = env.Describe("IMGPROXY_CLOUD_WATCH_NAMESPACE", "string")
	IMGPROXY_CLOUD_WATCH_REGION       = env.Describe("IMGPROXY_CLOUD_WATCH_REGION", "string")
)

// Config holds the configuration for CloudWatch monitoring
type Config struct {
	ServiceName     string        // CloudWatch service name (also used to enable/disable CloudWatch)
	Namespace       string        // CloudWatch metrics namespace
	Region          string        // AWS region for CloudWatch
	MetricsInterval time.Duration // Interval between metrics collections
}

// NewDefaultConfig returns a new default configuration for CloudWatch monitoring
func NewDefaultConfig() Config {
	return Config{
		ServiceName:     "",         // CloudWatch service name, enabled if not empty
		Namespace:       "imgproxy", // CloudWatch metrics namespace
		Region:          "",
		MetricsInterval: 10 * time.Second, // Metrics collection interval
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.ServiceName, IMGPROXY_CLOUD_WATCH_SERVICE_NAME),
		env.String(&c.Namespace, IMGPROXY_CLOUD_WATCH_NAMESPACE),
		env.String(&c.Region, IMGPROXY_CLOUD_WATCH_REGION),
	)

	return c, err
}

// Enabled returns true if CloudWatch is enabled
func (c *Config) Enabled() bool {
	return len(c.ServiceName) > 0
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	// If service name is not set, CloudWatch is disabled, so no need to validate other fields
	if !c.Enabled() {
		return nil
	}

	return nil
}
