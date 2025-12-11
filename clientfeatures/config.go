package clientfeatures

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_AUTO_WEBP           = env.Bool("IMGPROXY_AUTO_WEBP")
	IMGPROXY_AUTO_AVIF           = env.Bool("IMGPROXY_AUTO_AVIF")
	IMGPROXY_AUTO_JXL            = env.Bool("IMGPROXY_AUTO_JXL")
	IMGPROXY_ENFORCE_WEBP        = env.Bool("IMGPROXY_ENFORCE_WEBP")
	IMGPROXY_ENFORCE_AVIF        = env.Bool("IMGPROXY_ENFORCE_AVIF")
	IMGPROXY_ENFORCE_JXL         = env.Bool("IMGPROXY_ENFORCE_JXL")
	IMGPROXY_ENABLE_CLIENT_HINTS = env.Bool("IMGPROXY_ENABLE_CLIENT_HINTS")
)

// Config holds configuration for response writer
type Config struct {
	AutoWebp    bool // Whether to automatically serve WebP when supported
	EnforceWebp bool // Whether to enforce WebP format
	AutoAvif    bool // Whether to automatically serve AVIF when supported
	EnforceAvif bool // Whether to enforce AVIF format
	AutoJxl     bool // Whether to automatically serve JXL when supported
	EnforceJxl  bool // Whether to enforce JXL format

	EnableClientHints bool // Whether to enable client hints support
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{} // All features disabled by default
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_AUTO_WEBP.Parse(&c.AutoWebp),
		IMGPROXY_ENFORCE_WEBP.Parse(&c.EnforceWebp),
		IMGPROXY_AUTO_AVIF.Parse(&c.AutoAvif),
		IMGPROXY_ENFORCE_AVIF.Parse(&c.EnforceAvif),
		IMGPROXY_AUTO_JXL.Parse(&c.AutoJxl),
		IMGPROXY_ENFORCE_JXL.Parse(&c.EnforceJxl),
	)

	return c, err
}
