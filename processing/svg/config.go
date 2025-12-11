package svg

import (
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_SANITIZE_SVG = env.Bool("IMGPROXY_SANITIZE_SVG")
)

// Config holds SVG-specific configuration
type Config struct {
	Sanitize bool // Sanitize SVG content for security
}

// NewDefaultConfig creates a new Config instance with default values
func NewDefaultConfig() Config {
	return Config{
		Sanitize: true, // By default, sanitize SVG for security
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := IMGPROXY_SANITIZE_SVG.Parse(&c.Sanitize)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	return nil
}
