package workers

import (
	"errors"
	"log/slog"
	"runtime"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_REQUESTS_QUEUE_SIZE = env.Int("IMGPROXY_REQUESTS_QUEUE_SIZE")
	IMGPROXY_WORKERS             = env.Int("IMGPROXY_WORKERS")
	AWS_LAMBDA_FUNCTION_NAME     = env.String("AWS_LAMBDA_FUNCTION_NAME")
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

	// AWS Lambda environment detected: no queues and single worker
	if fnName, _ := AWS_LAMBDA_FUNCTION_NAME.GetEnv(); len(fnName) > 0 {
		c.WorkersNumber = 1
		c.RequestsQueueSize = 0

		slog.Info("AWS Lambda environment detected, setting workers to 1")

		return c, nil
	}

	err := errors.Join(
		IMGPROXY_REQUESTS_QUEUE_SIZE.Parse(&c.RequestsQueueSize),
		IMGPROXY_WORKERS.Parse(&c.WorkersNumber),
	)

	return c, err
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.RequestsQueueSize < 0 {
		return IMGPROXY_REQUESTS_QUEUE_SIZE.ErrorNegative()
	}

	if c.WorkersNumber <= 0 {
		return IMGPROXY_WORKERS.ErrorZeroOrNegative()
	}

	return nil
}
