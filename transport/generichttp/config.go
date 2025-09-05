package generichttp

import (
	"fmt"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

// Config holds the configuration for the generic HTTP transport
type Config struct {
	ClientKeepAliveTimeout time.Duration
	IgnoreSslVerification  bool
}

// NewDefaultConfig returns a new default configuration for the generic HTTP transport
func NewDefaultConfig() *Config {
	return &Config{
		ClientKeepAliveTimeout: 90 * time.Second,
		IgnoreSslVerification:  false,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c.ClientKeepAliveTimeout = time.Duration(config.ClientKeepAliveTimeout) * time.Second
	c.IgnoreSslVerification = config.IgnoreSslVerification

	return c, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if c.ClientKeepAliveTimeout < 0 {
		return fmt.Errorf("client KeepAlive timeout should be greater than or equal to 0, now - %d", c.ClientKeepAliveTimeout)
	}

	return nil
}
