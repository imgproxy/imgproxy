package gcs

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions
// for Google Cloud Storage storage
type ConfigDesc struct {
	Key            env.StringVar
	Endpoint       env.StringVar
	AllowedBuckets env.StringSliceVar
	DeniedBuckets  env.StringSliceVar
}

// Config holds the configuration for Google Cloud Storage transport
type Config struct {
	Key            string   // Google Cloud Storage service account key
	Endpoint       string   // Google Cloud Storage endpoint URL
	ReadOnly       bool     // Read-only access
	AllowedBuckets []string // List of allowed buckets
	DeniedBuckets  []string // List of denied buckets
	TestNoAuth     bool     // disable authentication for tests
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
		TestNoAuth:     false,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		desc.Key.Parse(&c.Key),
		desc.Endpoint.Parse(&c.Endpoint),
		desc.AllowedBuckets.Parse(&c.AllowedBuckets),
		desc.DeniedBuckets.Parse(&c.DeniedBuckets),
	)

	c.desc = desc

	return c, err
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	return nil
}
