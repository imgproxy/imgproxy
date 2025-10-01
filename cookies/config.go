package cookies

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_COOKIE_PASSTHROUGH     = env.Describe("IMGPROXY_COOKIE_PASSTHROUGH", "boolean")
	IMGPROXY_COOKIE_PASSTHROUGH_ALL = env.Describe("IMGPROXY_COOKIE_PASSTHROUGH_ALL", "boolean")
	IMGPROXY_COOKIE_BASE_URL        = env.Describe("IMGPROXY_COOKIE_BASE_URL", "string")
)

// Config holds cookie-related configuration.
type Config struct {
	CookiePassthrough    bool
	CookiePassthroughAll bool
	CookieBaseURL        string
}

// NewDefaultConfig creates a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		CookiePassthroughAll: false,
		CookieBaseURL:        "",
		CookiePassthrough:    false,
	}
}

// LoadConfigFromEnv creates a new Config instance loading values from environment variables.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.Bool(&c.CookiePassthrough, IMGPROXY_COOKIE_PASSTHROUGH),
		env.Bool(&c.CookiePassthroughAll, IMGPROXY_COOKIE_PASSTHROUGH_ALL),
		env.String(&c.CookieBaseURL, IMGPROXY_COOKIE_BASE_URL),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// No validation needed for cookie config currently
	return nil
}
