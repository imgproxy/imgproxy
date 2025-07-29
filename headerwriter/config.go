package headerwriter

import (
	"github.com/imgproxy/imgproxy/v3/config"
)

// Config is the package-local configuration
type Config struct {
	// SetCanonicalHeader indicates whether to set the canonical header
	SetCanonicalHeader bool

	// TTL is the default Cache-Control max-age= value for cached images
	DefaultTTL int

	// CacheControlPassthrough indicates whether to passthrough the Cache-Control header
	// from the original response
	CacheControlPassthrough bool

	// LastModifiedEnabled indicates whether to set the Last-Modified header
	LastModifiedEnabled bool

	// EnableClientHints indicates whether to enable Client Hints in Vary header
	EnableClientHints bool

	// SetVaryAccept indicates that the Vary header should include Accept
	SetVaryAccept bool
}

// NewConfigFromEnv creates a new Config instance from the current configuration
func NewConfigFromEnv() *Config {
	return &Config{
		SetCanonicalHeader:      config.SetCanonicalHeader,
		DefaultTTL:              config.TTL,
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
