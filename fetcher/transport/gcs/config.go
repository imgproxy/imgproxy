package gcs

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_GCS_KEY      = env.Describe("IMGPROXY_GCS_KEY", "string")
	IMGPROXY_GCS_ENDPOINT = env.Describe("IMGPROXY_GCS_ENDPOINT", "string")
)

// Config holds the configuration for Google Cloud Storage transport
type Config struct {
	Key      string // Google Cloud Storage service account key
	Endpoint string // Google Cloud Storage endpoint URL
}

// NewDefaultConfig returns a new default configuration for Google Cloud Storage transport
func NewDefaultConfig() Config {
	return Config{
		Key:      "",
		Endpoint: "",
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Key, IMGPROXY_GCS_KEY),
		env.String(&c.Endpoint, IMGPROXY_GCS_ENDPOINT),
	)

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
