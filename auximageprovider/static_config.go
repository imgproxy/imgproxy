package auximageprovider

import (
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_WATERMARK_DATA = env.Describe("IMGPROXY_WATERMARK_DATA", "base64-encoded string")
	IMGPROXY_WATERMARK_PATH = env.Describe("IMGPROXY_WATERMARK_PATH", "path")
	IMGPROXY_WATERMARK_URL  = env.Describe("IMGPROXY_WATERMARK_URL", "URL")

	IMGPROXY_FALLBACK_IMAGE_DATA = env.Describe("IMGPROXY_FALLBACK_IMAGE_DATA", "base64-encoded string")
	IMGPROXY_FALLBACK_IMAGE_PATH = env.Describe("IMGPROXY_FALLBACK_IMAGE_PATH", "path")
	IMGPROXY_FALLBACK_IMAGE_URL  = env.Describe("IMGPROXY_FALLBACK_IMAGE_URL", "URL")
)

// StaticConfig holds the configuration for the auxiliary image provider
type StaticConfig struct {
	Base64Data string
	Path       string
	URL        string
}

// NewDefaultStaticConfig creates a new default configuration for the auxiliary image provider
func NewDefaultStaticConfig() StaticConfig {
	return StaticConfig{
		Base64Data: "",
		Path:       "",
		URL:        "",
	}
}

// LoadWatermarkStaticConfigFromEnv loads the watermark configuration from the environment
func LoadWatermarkStaticConfigFromEnv(c *StaticConfig) (*StaticConfig, error) {
	c = ensure.Ensure(c, NewDefaultStaticConfig)

	env.String(&c.Base64Data, IMGPROXY_WATERMARK_DATA)
	env.String(&c.Path, IMGPROXY_WATERMARK_PATH)
	env.String(&c.URL, IMGPROXY_WATERMARK_URL)

	return c, nil
}

// LoadFallbackStaticConfigFromEnv loads the fallback configuration from the environment
func LoadFallbackStaticConfigFromEnv(c *StaticConfig) (*StaticConfig, error) {
	c = ensure.Ensure(c, NewDefaultStaticConfig)

	env.String(&c.Base64Data, IMGPROXY_FALLBACK_IMAGE_DATA)
	env.String(&c.Path, IMGPROXY_FALLBACK_IMAGE_PATH)
	env.String(&c.URL, IMGPROXY_FALLBACK_IMAGE_URL)

	return c, nil
}
