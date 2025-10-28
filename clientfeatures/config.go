package clientfeatures

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_AUTO_WEBP           = env.Describe("IMGPROXY_AUTO_WEBP", "boolean")
	IMGPROXY_AUTO_AVIF           = env.Describe("IMGPROXY_AUTO_AVIF", "boolean")
	IMGPROXY_AUTO_JXL            = env.Describe("IMGPROXY_AUTO_JXL", "boolean")
	IMGPROXY_ENFORCE_WEBP        = env.Describe("IMGPROXY_ENFORCE_WEBP", "boolean")
	IMGPROXY_ENFORCE_AVIF        = env.Describe("IMGPROXY_ENFORCE_AVIF", "boolean")
	IMGPROXY_ENFORCE_JXL         = env.Describe("IMGPROXY_ENFORCE_JXL", "boolean")
	IMGPROXY_ENABLE_CLIENT_HINTS = env.Describe("IMGPROXY_ENABLE_CLIENT_HINTS", "boolean")
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
		env.Bool(&c.AutoWebp, IMGPROXY_AUTO_WEBP),
		env.Bool(&c.EnforceWebp, IMGPROXY_ENFORCE_WEBP),
		env.Bool(&c.AutoAvif, IMGPROXY_AUTO_AVIF),
		env.Bool(&c.EnforceAvif, IMGPROXY_ENFORCE_AVIF),
		env.Bool(&c.AutoJxl, IMGPROXY_AUTO_JXL),
		env.Bool(&c.EnforceJxl, IMGPROXY_ENFORCE_JXL),
	)

	return c, err
}
