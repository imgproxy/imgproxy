package gcs

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions
// for Google Cloud Storage storage
type ConfigDesc struct {
	Key            env.Desc
	Endpoint       env.Desc
	AllowedBuckets env.Desc
	DeniedBuckets  env.Desc
}

// Config holds the configuration for Google Cloud Storage transport
type Config struct {
	Key            string   // Google Cloud Storage service account key
	Endpoint       string   // Google Cloud Storage endpoint URL
	ReadOnly       bool     // Read-only access
	AllowedBuckets []string // List of allowed buckets
	DeniedBuckets  []string // List of denied buckets
	desc           ConfigDesc
}

// NewDefaultConfig returns a new default configuration for Google Cloud Storage transport
func NewDefaultConfig() Config {
	return Config{
		Key:            "",
		Endpoint:       "",
		ReadOnly:       true,
		AllowedBuckets: nil,
		DeniedBuckets:  nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.String(&c.Key, desc.Key),
		env.String(&c.Endpoint, desc.Endpoint),
		env.StringSlice(&c.AllowedBuckets, desc.AllowedBuckets),
		env.StringSlice(&c.DeniedBuckets, desc.DeniedBuckets),
	)

	c.desc = desc

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
