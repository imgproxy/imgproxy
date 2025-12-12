package abs

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions for
// Azure Blob Storage transport
type ConfigDesc struct {
	Name           env.StringVar
	Endpoint       env.StringVar
	Key            env.StringVar
	AllowedBuckets env.StringSliceVar
	DeniedBuckets  env.StringSliceVar
}

// Config holds the configuration for Azure Blob Storage transport
type Config struct {
	Name           string   // Azure storage account name
	Endpoint       string   // Azure Blob Storage endpoint URL
	Key            string   // Azure storage account key
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
		AllowedBuckets: nil,
		DeniedBuckets:  nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		desc.Name.Parse(&c.Name),
		desc.Endpoint.Parse(&c.Endpoint),
		desc.Key.Parse(&c.Key),
		desc.AllowedBuckets.Parse(&c.AllowedBuckets),
		desc.DeniedBuckets.Parse(&c.DeniedBuckets),
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
