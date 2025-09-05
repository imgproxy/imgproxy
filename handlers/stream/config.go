package stream

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Config represents the configuration for the image streamer
type Config struct {
	// CookiePassthrough indicates whether cookies should be passed through to the image response
	CookiePassthrough bool

	// PassthroughRequestHeaders specifies the request headers to include in the passthrough response
	PassthroughRequestHeaders []string

	// PassthroughResponseHeaders specifies the response headers to copy from the response
	PassthroughResponseHeaders []string
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		CookiePassthrough: false,
		PassthroughRequestHeaders: []string{
			httpheaders.IfNoneMatch,
			httpheaders.IfModifiedSince,
			httpheaders.AcceptEncoding,
			httpheaders.Range,
		},
		PassthroughResponseHeaders: []string{
			httpheaders.ContentType,
			httpheaders.ContentEncoding,
			httpheaders.ContentRange,
			httpheaders.AcceptRanges,
			httpheaders.LastModified,
			httpheaders.Etag,
		},
	}
}

// LoadConfigFromEnv loads config variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.CookiePassthrough = config.CookiePassthrough

	return c, nil
}

// Validate checks config for errors
func (c *Config) Validate() error {
	return nil
}
