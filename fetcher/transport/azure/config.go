package azure

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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

	c.Name = config.ABSName
	c.Endpoint = config.ABSEndpoint
	c.Key = config.ABSKey

	return c, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Name) == 0 {
		return fmt.Errorf("azure account name must be set")
	}

	return nil
}
