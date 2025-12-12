package fetcher

import (
	"errors"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport"
	"github.com/imgproxy/imgproxy/v3/version"
)

var (
	IMGPROXY_USER_AGENT       = env.String("IMGPROXY_USER_AGENT")
	IMGPROXY_DOWNLOAD_TIMEOUT = env.Duration("IMGPROXY_DOWNLOAD_TIMEOUT")
	IMGPROXY_MAX_REDIRECTS    = env.Int("IMGPROXY_MAX_REDIRECTS")
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

	_, trErr := transport.LoadConfigFromEnv(&c.Transport)

	err := errors.Join(
		trErr,
		IMGPROXY_USER_AGENT.Parse(&c.UserAgent),
		IMGPROXY_DOWNLOAD_TIMEOUT.Parse(&c.DownloadTimeout),
		IMGPROXY_MAX_REDIRECTS.Parse(&c.MaxRedirects),
	)

	// Set the current version in the User-Agent string
	c.UserAgent = strings.ReplaceAll(c.UserAgent, "%current_version", version.Version)

	return c, err
}

// Validate checks config for errors
func (c *Config) Validate() error {
	if len(c.UserAgent) == 0 {
		return IMGPROXY_USER_AGENT.ErrorEmpty()
	}

	if c.DownloadTimeout <= 0 {
		return IMGPROXY_DOWNLOAD_TIMEOUT.ErrorZeroOrNegative()
	}

	if c.MaxRedirects <= 0 {
		return IMGPROXY_MAX_REDIRECTS.ErrorZeroOrNegative()
	}

	return nil
}
