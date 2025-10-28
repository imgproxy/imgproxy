package responsewriter

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_SET_CANONICAL_HEADER      = env.Describe("IMGPROXY_SET_CANONICAL_HEADER", "boolean")
	IMGPROXY_TTL                       = env.Describe("IMGPROXY_TTL", "seconds >= 0")
	IMGPROXY_FALLBACK_IMAGE_TTL        = env.Describe("IMGPROXY_FALLBACK_IMAGE_TTL", "seconds >= 0")
	IMGPROXY_CACHE_CONTROL_PASSTHROUGH = env.Describe("IMGPROXY_CACHE_CONTROL_PASSTHROUGH", "boolean")
	IMGPROXY_WRITE_RESPONSE_TIMEOUT    = env.Describe("IMGPROXY_WRITE_RESPONSE_TIMEOUT", "seconds > 0")
)

// Config holds configuration for response writer
type Config struct {
	SetCanonicalHeader      bool          // Indicates whether to set the canonical header
	DefaultTTL              int           // Default Cache-Control max-age= value for cached images
	FallbackImageTTL        int           // TTL for images served as fallbacks
	CacheControlPassthrough bool          // Passthrough the Cache-Control from the original response
	WriteResponseTimeout    time.Duration // Timeout for response write operations
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		SetCanonicalHeader:      false,
		DefaultTTL:              31_536_000,
		FallbackImageTTL:        0,
		CacheControlPassthrough: false,
		WriteResponseTimeout:    10 * time.Second,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.Bool(&c.SetCanonicalHeader, IMGPROXY_SET_CANONICAL_HEADER),
		env.Int(&c.DefaultTTL, IMGPROXY_TTL),
		env.Int(&c.FallbackImageTTL, IMGPROXY_FALLBACK_IMAGE_TTL),
		env.Bool(&c.CacheControlPassthrough, IMGPROXY_CACHE_CONTROL_PASSTHROUGH),
		env.Duration(&c.WriteResponseTimeout, IMGPROXY_WRITE_RESPONSE_TIMEOUT),
	)

	return c, err
}

// Validate checks config for errors
func (c *Config) Validate() error {
	if c.DefaultTTL < 0 {
		return IMGPROXY_TTL.ErrorNegative()
	}

	if c.FallbackImageTTL < 0 {
		return IMGPROXY_FALLBACK_IMAGE_TTL.ErrorNegative()
	}

	if c.WriteResponseTimeout <= 0 {
		return IMGPROXY_WRITE_RESPONSE_TIMEOUT.ErrorZeroOrNegative()
	}

	return nil
}
