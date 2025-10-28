package azure

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions for
// Azure Blob Storage transport
type ConfigDesc struct {
	Name           env.Desc
	Endpoint       env.Desc
	Key            env.Desc
	AllowedBuckets env.Desc
	DeniedBuckets  env.Desc
}

// Config holds the configuration for Azure Blob Storage transport
type Config struct {
	Name           string   // Azure storage account name
	Endpoint       string   // Azure Blob Storage endpoint URL
	Key            string   // Azure storage account key
	ReadOnly       bool     // Read-only access
	AllowedBuckets []string // List of allowed buckets (containers)
	DeniedBuckets  []string // List of denied buckets (containers)
	desc           ConfigDesc
}

// NewDefaultConfig returns a new default configuration for Azure Blob Storage transport
func NewDefaultConfig() Config {
	return Config{
		Name:           "",
		Endpoint:       "",
		Key:            "",
		ReadOnly:       true,
		AllowedBuckets: nil,
		DeniedBuckets:  nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Name, desc.Name),
		env.String(&c.Endpoint, desc.Endpoint),
		env.String(&c.Key, desc.Key),
		env.StringSlice(&c.AllowedBuckets, desc.AllowedBuckets),
		env.StringSlice(&c.DeniedBuckets, desc.DeniedBuckets),
	)

	c.desc = desc

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Name) == 0 {
		return c.desc.Name.ErrorEmpty()
	}

	return nil
}
