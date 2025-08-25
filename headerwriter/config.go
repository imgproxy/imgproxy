package headerwriter

import (
	"github.com/imgproxy/imgproxy/v3/config"
)

// Config is the package-local configuration
type Config struct {
	SetCanonicalHeader      bool // Indicates whether to set the canonical header
	DefaultTTL              int  // Default Cache-Control max-age= value for cached images
	FallbackImageTTL        int  // TTL for images served as fallbacks
	CacheControlPassthrough bool // Passthrough the Cache-Control from the original response
	LastModifiedEnabled     bool // Set the Last-Modified header
	EnableClientHints       bool // Enable Vary header
	SetVaryAccept           bool // Whether to include Accept in Vary header
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() *Config {
	return &Config{
		SetCanonicalHeader:      false,
		DefaultTTL:              31536000,
		FallbackImageTTL:        0,
		LastModifiedEnabled:     false,
		CacheControlPassthrough: false,
		EnableClientHints:       false,
		SetVaryAccept:           false,
	}
}

// LoadFromEnv overrides configuration variables from environment
func (c *Config) LoadFromEnv() *Config {
	c.SetCanonicalHeader = config.SetCanonicalHeader
	c.DefaultTTL = config.TTL
	c.FallbackImageTTL = config.FallbackImageTTL
	c.LastModifiedEnabled = config.LastModifiedEnabled
	c.CacheControlPassthrough = config.CacheControlPassthrough
	c.EnableClientHints = config.EnableClientHints
	c.SetVaryAccept = config.AutoWebp ||
		config.EnforceWebp ||
		config.AutoAvif ||
		config.EnforceAvif ||
		config.AutoJxl ||
		config.EnforceJxl

	return c
}

// NewConfigFromEnv creates a new Config instance from the current configuration
func NewConfigFromEnv() *Config {
	return NewDefaultConfig().LoadFromEnv()
}
