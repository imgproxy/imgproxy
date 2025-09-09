package fetcher

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport"
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

	// Transport holds the configuration for the transport layer.
	Transport transport.Config
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		UserAgent:       "imgproxy/" + version.Version,
		DownloadTimeout: 5 * time.Second,
		MaxRedirects:    10,
		Transport:       transport.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv loads config variables from env
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.UserAgent = config.UserAgent
	c.DownloadTimeout = time.Duration(config.DownloadTimeout) * time.Second
	c.MaxRedirects = config.MaxRedirects

	_, err := transport.LoadConfigFromEnv(&c.Transport)
	if err != nil {
		return nil, err
	}

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
