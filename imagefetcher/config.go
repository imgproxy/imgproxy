package imagefetcher

import "github.com/imgproxy/imgproxy/v3/config"

type Config struct {
	// MaxRedirects is the maximum number of redirects allowed when fetching images.
	MaxRedirects int
}

func NewConfigFromEnv() *Config {
	return &Config{
		MaxRedirects: config.MaxRedirects,
	}
}
