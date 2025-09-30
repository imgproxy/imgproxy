package processing

import (
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_COOKIE_PASSTHROUGH        = env.Describe("IMGPROXY_COOKIE_PASSTHROUGH", "boolean")
	IMGPROXY_REPORT_DOWNLOADING_ERRORS = env.Describe("IMGPROXY_REPORT_DOWNLOADING_ERRORS", "boolean")
	IMGPROXY_LAST_MODIFIED_ENABLED     = env.Describe("IMGPROXY_LAST_MODIFIED_ENABLED", "boolean")
	IMGPROXY_ETAG_ENABLED              = env.Describe("IMGPROXY_ETAG_ENABLED", "boolean")
	IMGPROXY_REPORT_IO_ERRORS          = env.Describe("IMGPROXY_REPORT_IO_ERRORS", "boolean")
	IMGPROXY_FALLBACK_IMAGE_HTTP_CODE  = env.Describe("IMGPROXY_FALLBACK_IMAGE_HTTP_CODE", "HTTP code")
	IMGPROXY_ENABLE_DEBUG_HEADERS      = env.Describe("IMGPROXY_ENABLE_DEBUG_HEADERS", "boolean")
)

// Config represents handler config
type Config struct {
	CookiePassthrough       bool // Whether to passthrough cookies
	ReportDownloadingErrors bool // Whether to report downloading errors
	LastModifiedEnabled     bool // Whether to enable Last-Modified
	ETagEnabled             bool // Whether to enable ETag
	ReportIOErrors          bool // Whether to report IO errors
	FallbackImageHTTPCode   int  // Fallback image HTTP status code
	EnableDebugHeaders      bool // Whether to enable debug headers
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		CookiePassthrough:       false,
		ReportDownloadingErrors: true,
		LastModifiedEnabled:     true,
		ETagEnabled:             true,
		ReportIOErrors:          false,
		FallbackImageHTTPCode:   http.StatusOK,
		EnableDebugHeaders:      false,
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.Bool(&c.CookiePassthrough, IMGPROXY_COOKIE_PASSTHROUGH),
		env.Bool(&c.ReportDownloadingErrors, IMGPROXY_REPORT_DOWNLOADING_ERRORS),
		env.Bool(&c.LastModifiedEnabled, IMGPROXY_LAST_MODIFIED_ENABLED),
		env.Bool(&c.ETagEnabled, IMGPROXY_ETAG_ENABLED),
		env.Bool(&c.ReportIOErrors, IMGPROXY_REPORT_IO_ERRORS),
		env.Int(&c.FallbackImageHTTPCode, IMGPROXY_FALLBACK_IMAGE_HTTP_CODE),
		env.Bool(&c.EnableDebugHeaders, IMGPROXY_ENABLE_DEBUG_HEADERS),
	)

	return c, err
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.FallbackImageHTTPCode != 0 && (c.FallbackImageHTTPCode < 100 || c.FallbackImageHTTPCode > 599) {
		return IMGPROXY_FALLBACK_IMAGE_HTTP_CODE.Errorf("invalid")
	}

	return nil
}
