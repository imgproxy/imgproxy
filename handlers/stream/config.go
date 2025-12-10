package stream

import (
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

var (
	// Processing handler has the similar variable.
	// For now, we do not want to couple hanlders/processing and handlers/stream packages,
	// so we duplicate it here. Discuss.
	//nolint:godoclint
	IMGPROXY_COOKIE_PASSTHROUGH = env.Describe("IMGPROXY_COOKIE_PASSTHROUGH", "boolean")
)

// Config represents the configuration for the image streamer.
type Config struct {
	// PassthroughRequestHeaders specifies the request headers to include in the passthrough response
	PassthroughRequestHeaders []string

	// PassthroughResponseHeaders specifies the response headers to copy from the response
	PassthroughResponseHeaders []string
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
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

// LoadConfigFromEnv loads config variables from environment.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)
	return c, nil
}

// Validate checks config for errors.
func (c *Config) Validate() error {
	return nil
}
