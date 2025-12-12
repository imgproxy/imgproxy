package workers

import (
	"errors"
	"runtime"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_REQUESTS_QUEUE_SIZE = env.Int("IMGPROXY_REQUESTS_QUEUE_SIZE")
	IMGPROXY_WORKERS_NUMBER      = env.Int("IMGPROXY_WORKERS_NUMBER")
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

	err := errors.Join(
		IMGPROXY_REQUESTS_QUEUE_SIZE.Parse(&c.RequestsQueueSize),
		IMGPROXY_WORKERS_NUMBER.Parse(&c.WorkersNumber),
	)

	return c, err
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.RequestsQueueSize < 0 {
		return IMGPROXY_REQUESTS_QUEUE_SIZE.ErrorNegative()
	}

	if c.WorkersNumber <= 0 {
		return IMGPROXY_WORKERS_NUMBER.ErrorZeroOrNegative()
	}

	return nil
}
