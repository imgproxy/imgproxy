package responsewriter

import (
	"fmt"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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
		DefaultTTL:              31536000,
		FallbackImageTTL:        0,
		CacheControlPassthrough: false,
		VaryValue:               "",
		WriteResponseTimeout:    10 * time.Second,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.SetCanonicalHeader = config.SetCanonicalHeader
	c.DefaultTTL = config.TTL
	c.FallbackImageTTL = config.FallbackImageTTL
	c.CacheControlPassthrough = config.CacheControlPassthrough
	c.WriteResponseTimeout = time.Duration(config.WriteResponseTimeout) * time.Second

	vary := make([]string, 0)

	if c.envEnableFormatDetection() {
		vary = append(vary, "Accept")
	}

	if c.envEnableClientHints() {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	c.VaryValue = strings.Join(vary, ", ")

	return c, nil
}

func (c *Config) envEnableFormatDetection() bool {
	return config.AutoWebp ||
		config.EnforceWebp ||
		config.AutoAvif ||
		config.EnforceAvif ||
		config.AutoJxl ||
		config.EnforceJxl
}

func (c *Config) envEnableClientHints() bool {
	return config.EnableClientHints
}

// Validate checks config for errors
func (c *Config) Validate() error {
	if c.DefaultTTL < 0 {
		return fmt.Errorf("image TTL should be greater than or equal to 0, now - %d", c.DefaultTTL)
	}

	if c.FallbackImageTTL < 0 {
		return fmt.Errorf("fallback image TTL should be greater than or equal to 0, now - %d", c.FallbackImageTTL)
	}

	if c.WriteResponseTimeout <= 0 {
		return fmt.Errorf("write response timeout should be greater than 0, now - %d", c.WriteResponseTimeout)
	}

	return nil
}
