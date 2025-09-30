package azure

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_ABS_NAME     = env.Describe("IMGPROXY_ABS_NAME", "string")
	IMGPROXY_ABS_ENDPOINT = env.Describe("IMGPROXY_ABS_ENDPOINT", "string")
	IMGPROXY_ABS_KEY      = env.Describe("IMGPROXY_ABS_KEY", "string")
)

// Config holds the configuration for Azure Blob Storage transport
type Config struct {
	Name     string // Azure storage account name
	Endpoint string // Azure Blob Storage endpoint URL
	Key      string // Azure storage account key
}

// NewDefaultConfig returns a new default configuration for Azure Blob Storage transport
func NewDefaultConfig() Config {
	return Config{
		Name:     "",
		Endpoint: "",
		Key:      "",
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Name, IMGPROXY_ABS_NAME),
		env.String(&c.Endpoint, IMGPROXY_ABS_ENDPOINT),
		env.String(&c.Key, IMGPROXY_ABS_KEY),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Name) == 0 {
		return IMGPROXY_ABS_NAME.ErrorEmpty()
	}

	return nil
}
