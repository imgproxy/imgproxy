package swift

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

// ConfigDesc holds the configuration descriptions for Swift storage
type ConfigDesc struct {
	Username       env.Desc
	APIKey         env.Desc
	AuthURL        env.Desc
	Domain         env.Desc
	Tenant         env.Desc
	AuthVersion    env.Desc
	ConnectTimeout env.Desc
	Timeout        env.Desc
	AllowedBuckets env.Desc
	DeniedBuckets  env.Desc
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
	ReadOnly       bool          // Read-only access
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
		env.String(&c.Username, desc.Username),
		env.String(&c.APIKey, desc.APIKey),
		env.String(&c.AuthURL, desc.AuthURL),
		env.String(&c.Domain, desc.Domain),
		env.String(&c.Tenant, desc.Tenant),
		env.Int(&c.AuthVersion, desc.AuthVersion),
		env.Duration(&c.ConnectTimeout, desc.ConnectTimeout),
		env.Duration(&c.Timeout, desc.Timeout),
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
