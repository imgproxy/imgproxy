package imagefetcher

import "github.com/imgproxy/imgproxy/v3/config"

// Config holds the configuration for the image fetcher.
type Config struct {
	// MaxRedirects is the maximum number of redirects to follow when fetching an image.
	MaxRedirects int
}

// NewConfigFromEnv creates a new Config instance from environment variables or defaults.
func NewConfigFromEnv() *Config {
	return &Config{
		MaxRedirects: config.MaxRedirects,
	}
}
