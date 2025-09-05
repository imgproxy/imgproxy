package headerwriter

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// Config is the package-local configuration
type Config struct {
	SetCanonicalHeader      bool // Indicates whether to set the canonical header
	DefaultTTL              int  // Default Cache-Control max-age= value for cached images
	FallbackImageTTL        int  // TTL for images served as fallbacks
	CacheControlPassthrough bool // Passthrough the Cache-Control from the original response
	EnableClientHints       bool // Enable Vary header
	SetVaryAccept           bool // Whether to include Accept in Vary header
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		SetCanonicalHeader:      false,
		DefaultTTL:              31536000,
		FallbackImageTTL:        0,
		CacheControlPassthrough: false,
		EnableClientHints:       false,
		SetVaryAccept:           false,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.SetCanonicalHeader = config.SetCanonicalHeader
	c.DefaultTTL = config.TTL
	c.FallbackImageTTL = config.FallbackImageTTL
	c.CacheControlPassthrough = config.CacheControlPassthrough
	c.EnableClientHints = config.EnableClientHints
	c.SetVaryAccept = config.AutoWebp ||
		config.EnforceWebp ||
		config.AutoAvif ||
		config.EnforceAvif ||
		config.AutoJxl ||
		config.EnforceJxl

	return c, nil
}

// Validate checks config for errors
func (c *Config) Validate() error {
	if c.DefaultTTL < 0 {
		return fmt.Errorf("image TTL should be greater than or equal to 0, now - %d", c.DefaultTTL)
	}

	if c.FallbackImageTTL < 0 {
		return fmt.Errorf("fallback image TTL should be greater than or equal to 0, now - %d", c.FallbackImageTTL)
	}

	return nil
}
