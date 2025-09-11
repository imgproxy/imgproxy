package workers

import (
	"fmt"
	"runtime"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// Config represents [Workers] config
type Config struct {
	RequestsQueueSize int // Maximum request queue size
	WorkersNumber     int // Number of allowed workers
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		RequestsQueueSize: 0,
		WorkersNumber:     runtime.GOMAXPROCS(0) * 2,
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.RequestsQueueSize = config.RequestsQueueSize
	c.WorkersNumber = config.Workers

	return c, nil
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.RequestsQueueSize < 0 {
		return fmt.Errorf("requests queue size should be greater than or equal 0, now - %d", c.RequestsQueueSize)
	}

	if c.WorkersNumber <= 0 {
		return fmt.Errorf("workers number should be greater than 0, now - %d", c.WorkersNumber)
	}

	return nil
}
