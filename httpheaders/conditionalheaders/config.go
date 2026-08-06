package conditionalheaders

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/env"
)

var (
	IMGPROXY_USE_LAST_MODIFIED    = env.Bool("IMGPROXY_USE_LAST_MODIFIED")
	IMGPROXY_LAST_MODIFIED_BUSTER = env.DateTime("IMGPROXY_LAST_MODIFIED_BUSTER")
	IMGPROXY_USE_ETAG             = env.Bool("IMGPROXY_USE_ETAG")
	IMGPROXY_ETAG_BUSTER          = env.String("IMGPROXY_ETAG_BUSTER")
)

// Config represents conditional headers config
type Config struct {
	LastModifiedEnabled bool      // Whether to enable Last-Modified
	LastModifiedBuster  time.Time // Last-Modified buster
	ETagEnabled         bool      // Whether to enable ETag
	ETagBuster          string    // ETag buster
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		LastModifiedEnabled: true,
		LastModifiedBuster:  time.Time{},
		ETagEnabled:         true,
		ETagBuster:          "",
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_USE_LAST_MODIFIED.Parse(&c.LastModifiedEnabled),
		IMGPROXY_USE_ETAG.Parse(&c.ETagEnabled),
		IMGPROXY_LAST_MODIFIED_BUSTER.Parse(&c.LastModifiedBuster),
		IMGPROXY_ETAG_BUSTER.Parse(&c.ETagBuster),
	)

	return c, err
}

// Validate checks configuration values
func (c *Config) Validate() error {
	return nil
}
