package stream

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Config represents the configuration for the image streamer
type Config struct {
	// CookiePassthrough indicates whether cookies should be passed through to the image response
	CookiePassthrough bool

	// PassthroughRequestHeaders specifies the request headers to include in the passthrough response
	PassthroughRequestHeaders []string

	// KeepResponseHeaders specifies the response headers to copy from the response
	KeepResponseHeaders []string
}

// NewConfigFromEnv creates a new Config instance from environment variables
func NewConfigFromEnv() *Config {
	return &Config{
		CookiePassthrough: config.CookiePassthrough,
		PassthroughRequestHeaders: []string{
			httpheaders.IfNoneMatch,
			httpheaders.IfModifiedSince,
			httpheaders.AcceptEncoding,
			httpheaders.Range,
		},
		KeepResponseHeaders: []string{
			httpheaders.ContentType,
			httpheaders.ContentEncoding,
			httpheaders.ContentRange,
			httpheaders.AcceptRanges,
			httpheaders.LastModified,
			httpheaders.Etag,
		},
	}
}
