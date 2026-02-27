package processing

import (
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_REPORT_DOWNLOADING_ERRORS = env.Bool("IMGPROXY_REPORT_DOWNLOADING_ERRORS")
	IMGPROXY_REPORT_IO_ERRORS          = env.Bool("IMGPROXY_REPORT_IO_ERRORS")
	IMGPROXY_FALLBACK_IMAGE_HTTP_CODE  = env.Int("IMGPROXY_FALLBACK_IMAGE_HTTP_CODE")
	IMGPROXY_ENABLE_DEBUG_HEADERS      = env.Bool("IMGPROXY_ENABLE_DEBUG_HEADERS")
)

// Config represents handler config
type Config struct {
	ReportDownloadingErrors bool // Whether to report downloading errors
	ReportIOErrors          bool // Whether to report IO errors
	FallbackImageHTTPCode   int  // Fallback image HTTP status code
	EnableDebugHeaders      bool // Whether to enable debug headers
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		ReportDownloadingErrors: true,
		ReportIOErrors:          false,
		FallbackImageHTTPCode:   http.StatusOK,
		EnableDebugHeaders:      false,
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_REPORT_DOWNLOADING_ERRORS.Parse(&c.ReportDownloadingErrors),
		IMGPROXY_REPORT_IO_ERRORS.Parse(&c.ReportIOErrors),
		IMGPROXY_FALLBACK_IMAGE_HTTP_CODE.Parse(&c.FallbackImageHTTPCode),
		IMGPROXY_ENABLE_DEBUG_HEADERS.Parse(&c.EnableDebugHeaders),
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
