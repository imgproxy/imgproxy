package imagefetcher

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/version"
)

// Config holds the configuration for the image fetcher.
type Config struct {
	// UserAgent is the User-Agent header to use when fetching images.
	UserAgent string
	// DownloadTimeout is the timeout for downloading an image, in seconds.
	DownloadTimeout time.Duration
	// MaxRedirects is the maximum number of redirects to follow when fetching an image.
	MaxRedirects int
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() *Config {
	return &Config{
		UserAgent:       "imgproxy/" + version.Version,
		DownloadTimeout: 5 * time.Second,
		MaxRedirects:    10,
	}
}

// LoadFromEnv loads config variables from env
func (c *Config) LoadFromEnv() (*Config, error) {
	c.UserAgent = config.UserAgent
	c.DownloadTimeout = time.Duration(config.DownloadTimeout) * time.Second
	c.MaxRedirects = config.MaxRedirects
	return c, nil
}

// Validate checks config for errors
func (c *Config) Validate() error {
	if len(c.UserAgent) == 0 {
		return errors.New("user agent cannot be empty")
	}

	if c.DownloadTimeout <= 0 {
		return errors.New("download timeout must be greater than 0")
	}

	return nil
}
