package swift

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions for Swift storage
type ConfigDesc struct {
	Username       env.StringVar
	APIKey         env.StringVar
	AuthURL        env.StringVar
	Domain         env.StringVar
	Tenant         env.StringVar
	AuthVersion    env.IntVar
	ConnectTimeout env.DurationVar
	Timeout        env.DurationVar
	AllowedBuckets env.StringSliceVar
	DeniedBuckets  env.StringSliceVar
}

// Config holds the configuration for Swift storage
type Config struct {
	Username       string        // Username for Swift authentication
	APIKey         string        // API key for Swift authentication
	AuthURL        string        // Authentication URL for Swift
	Domain         string        // Domain for Swift authentication
	Tenant         string        // Tenant for Swift authentication
	AuthVersion    int           // Authentication version for Swift
	ConnectTimeout time.Duration // Connection timeout for Swift
	Timeout        time.Duration // Request timeout for Swift
	AllowedBuckets []string      // List of allowed buckets (containers)
	DeniedBuckets  []string      // List of denied buckets (containers)
	desc           ConfigDesc
}

// NewDefaultConfig returns a new default configuration for Swift storage
func NewDefaultConfig() Config {
	return Config{
		Username:       "",
		APIKey:         "",
		AuthURL:        "",
		Domain:         "",
		Tenant:         "",
		AuthVersion:    0,
		ConnectTimeout: 10 * time.Second,
		Timeout:        60 * time.Second,
		AllowedBuckets: nil,
		DeniedBuckets:  nil,
	}
}

// LoadConfigFromEnv loads configuration from the global config package
func LoadConfigFromEnv(desc ConfigDesc, c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		desc.Username.Parse(&c.Username),
		desc.APIKey.Parse(&c.APIKey),
		desc.AuthURL.Parse(&c.AuthURL),
		desc.Domain.Parse(&c.Domain),
		desc.Tenant.Parse(&c.Tenant),
		desc.AuthVersion.Parse(&c.AuthVersion),
		desc.ConnectTimeout.Parse(&c.ConnectTimeout),
		desc.Timeout.Parse(&c.Timeout),
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
