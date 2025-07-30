package imagefetcher

import "github.com/imgproxy/imgproxy/v3/config"

type Config struct {
	// MaxRedirects is the maximum number of redirects to follow
	MaxRedirects int

	// UserAgent is the user agent string to use for requests
	UserAgent string

	// DownloadTimeout is the timeout for downloading images in seconds
	DownloadTimeout int
}

func NewConfigFromEnv() *Config {
	return &Config{
		MaxRedirects:    config.MaxRedirects,
		UserAgent:       config.UserAgent,
		DownloadTimeout: config.DownloadTimeout,
	}
}
