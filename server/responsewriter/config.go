package responsewriter

import (
	"errors"
	"strings"
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

	// NOTE: These are referenced here to determine if we need to set the Vary header
	// Unfotunately, we can not reuse them from optionsparser package due to import cycle
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
	SetCanonicalHeader      bool          // Indicates whether to set the canonical header
	DefaultTTL              int           // Default Cache-Control max-age= value for cached images
	FallbackImageTTL        int           // TTL for images served as fallbacks
	CacheControlPassthrough bool          // Passthrough the Cache-Control from the original response
	VaryValue               string        // Value for Vary header
	WriteResponseTimeout    time.Duration // Timeout for response write operations
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		SetCanonicalHeader:      false,
		DefaultTTL:              31_536_000,
		FallbackImageTTL:        0,
		CacheControlPassthrough: false,
		VaryValue:               "",
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
	if err != nil {
		return nil, err
	}

	vary := make([]string, 0)

	var ok bool

	if err, ok = c.envEnableFormatDetection(); err != nil {
		return nil, err
	}
	if ok {
		vary = append(vary, "Accept")
	}

	if err, ok = c.envEnableClientHints(); err != nil {
		return nil, err
	}
	if ok {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	c.VaryValue = strings.Join(vary, ", ")

	return c, nil
}

// envEnableFormatDetection checks if any of the format detection options are enabled
func (c *Config) envEnableFormatDetection() (error, bool) {
	var autoWebp, enforceWebp, autoAvif, enforceAvif, autoJxl, enforceJxl bool

	// We won't need those variables in runtime, hence, we could
	// read them here once into local variables
	err := errors.Join(
		env.Bool(&autoWebp, IMGPROXY_AUTO_WEBP),
		env.Bool(&enforceWebp, IMGPROXY_ENFORCE_WEBP),
		env.Bool(&autoAvif, IMGPROXY_AUTO_AVIF),
		env.Bool(&enforceAvif, IMGPROXY_ENFORCE_AVIF),
		env.Bool(&autoJxl, IMGPROXY_AUTO_JXL),
		env.Bool(&enforceJxl, IMGPROXY_ENFORCE_JXL),
	)
	if err != nil {
		return err, false
	}

	return nil, autoWebp ||
		enforceWebp ||
		autoAvif ||
		enforceAvif ||
		autoJxl ||
		enforceJxl
}

// envEnableClientHints checks if client hints are enabled
func (c *Config) envEnableClientHints() (err error, ok bool) {
	err = env.Bool(&ok, IMGPROXY_ENABLE_CLIENT_HINTS)
	return
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
