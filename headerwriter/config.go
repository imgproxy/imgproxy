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

// NewConfigFromEnv creates a new Config instance from the current configuration
func NewConfigFromEnv() *Config {
	return &Config{
		SetCanonicalHeader:      config.SetCanonicalHeader,
		DefaultTTL:              config.TTL,
		FallbackImageTTL:        config.FallbackImageTTL,
		LastModifiedEnabled:     config.LastModifiedEnabled,
		CacheControlPassthrough: config.CacheControlPassthrough,
		EnableClientHints:       config.EnableClientHints,
		SetVaryAccept: config.AutoWebp ||
			config.EnforceWebp ||
			config.AutoAvif ||
			config.EnforceAvif ||
			config.AutoJxl ||
			config.EnforceJxl,
	}
}
