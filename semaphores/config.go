package semaphores

import (
	"fmt"
	"runtime"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// Config represents handler config
type Config struct {
	RequestsQueueSize int // Request queue size
	Workers           int // Number of workers
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		RequestsQueueSize: 0,
		Workers:           runtime.GOMAXPROCS(0) * 2,
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.RequestsQueueSize = config.RequestsQueueSize
	c.Workers = config.Workers

	return c, nil
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.RequestsQueueSize < 0 {
		return fmt.Errorf("requests queue size should be greater than or equal 0, now - %d", c.RequestsQueueSize)
	}

	if c.Workers <= 0 {
		return fmt.Errorf("workers number should be greater than 0, now - %d", c.Workers)
	}

	return nil
}
