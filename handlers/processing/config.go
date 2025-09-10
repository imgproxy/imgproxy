package processing

import (
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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

	c.CookiePassthrough = config.CookiePassthrough
	c.ReportDownloadingErrors = config.ReportDownloadingErrors
	c.LastModifiedEnabled = config.LastModifiedEnabled
	c.ETagEnabled = config.ETagEnabled
	c.ReportIOErrors = config.ReportIOErrors
	c.FallbackImageHTTPCode = config.FallbackImageHTTPCode
	c.EnableDebugHeaders = config.EnableDebugHeaders

	return c, nil
}

// Validate checks configuration values
func (c *Config) Validate() error {
	if c.FallbackImageHTTPCode != 0 && (c.FallbackImageHTTPCode < 100 || c.FallbackImageHTTPCode > 599) {
		return errors.New("fallback image HTTP code should be between 100 and 599")
	}
	return nil
}
