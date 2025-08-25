package imagefetcher

import "github.com/imgproxy/imgproxy/v3/config"

// Config holds the configuration for the image fetcher.
type Config struct {
	// MaxRedirects is the maximum number of redirects to follow when fetching an image.
	MaxRedirects int
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() *Config {
	return &Config{
		MaxRedirects: 10,
	}
}

// LoadFromEnv loads config variables from env
func (c *Config) LoadFromEnv() *Config {
	c.MaxRedirects = config.MaxRedirects
	return c
}

// NewConfigFromEnv creates a new Config instance from environment variables or defaults.
func NewConfigFromEnv() *Config {
	return NewDefaultConfig().LoadFromEnv()
}
