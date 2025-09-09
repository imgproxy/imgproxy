package gcs

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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

	c.Key = config.GCSKey
	c.Endpoint = config.GCSEndpoint

	return c, nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
