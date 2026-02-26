package processing

import (
	"errors"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_REPORT_DOWNLOADING_ERRORS = env.Bool("IMGPROXY_REPORT_DOWNLOADING_ERRORS")
	IMGPROXY_LAST_MODIFIED_ENABLED     = env.Bool("IMGPROXY_LAST_MODIFIED_ENABLED")
	IMGPROXY_LAST_MODIFIED_BUSTER      = env.DateTime("IMGPROXY_LAST_MODIFIED_BUSTER")
	IMGPROXY_ETAG_ENABLED              = env.Bool("IMGPROXY_ETAG_ENABLED")
	IMGPROXY_ETAG_BUSTER               = env.String("IMGPROXY_ETAG_BUSTER")
	IMGPROXY_REPORT_IO_ERRORS          = env.Bool("IMGPROXY_REPORT_IO_ERRORS")
	IMGPROXY_FALLBACK_IMAGE_HTTP_CODE  = env.Int("IMGPROXY_FALLBACK_IMAGE_HTTP_CODE")
	IMGPROXY_ENABLE_DEBUG_HEADERS      = env.Bool("IMGPROXY_ENABLE_DEBUG_HEADERS")
)

// Config represents handler config
type Config struct {
	ReportDownloadingErrors bool   // Whether to report downloading errors
	LastModifiedEnabled     bool   // Whether to enable Last-Modified
	ETagEnabled             bool   // Whether to enable ETag
	ETagBuster              string // ETag buster
	ReportIOErrors          bool   // Whether to report IO errors
	FallbackImageHTTPCode   int    // Fallback image HTTP status code
	EnableDebugHeaders      bool   // Whether to enable debug headers
	LastModifiedBuster      time.Time
}

// NewDefaultConfig creates a new configuration with defaults
func NewDefaultConfig() Config {
	return Config{
		ReportDownloadingErrors: true,
		LastModifiedEnabled:     true,
		ETagEnabled:             true,
		ETagBuster:              "",
		ReportIOErrors:          false,
		FallbackImageHTTPCode:   http.StatusOK,
		EnableDebugHeaders:      false,
		LastModifiedBuster:      time.Time{},
	}
}

// LoadConfigFromEnv loads config from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_REPORT_DOWNLOADING_ERRORS.Parse(&c.ReportDownloadingErrors),
		IMGPROXY_LAST_MODIFIED_ENABLED.Parse(&c.LastModifiedEnabled),
		IMGPROXY_ETAG_ENABLED.Parse(&c.ETagEnabled),
		IMGPROXY_REPORT_IO_ERRORS.Parse(&c.ReportIOErrors),
		IMGPROXY_FALLBACK_IMAGE_HTTP_CODE.Parse(&c.FallbackImageHTTPCode),
		IMGPROXY_ENABLE_DEBUG_HEADERS.Parse(&c.EnableDebugHeaders),
		IMGPROXY_LAST_MODIFIED_BUSTER.Parse(&c.LastModifiedBuster),
		IMGPROXY_ETAG_BUSTER.Parse(&c.ETagBuster),
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
