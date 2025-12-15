package auximageprovider

import (
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_WATERMARK_DATA = env.String("IMGPROXY_WATERMARK_DATA")
	IMGPROXY_WATERMARK_PATH = env.String("IMGPROXY_WATERMARK_PATH")
	IMGPROXY_WATERMARK_URL  = env.String("IMGPROXY_WATERMARK_URL")

	IMGPROXY_FALLBACK_IMAGE_DATA = env.String("IMGPROXY_FALLBACK_IMAGE_DATA")
	IMGPROXY_FALLBACK_IMAGE_PATH = env.String("IMGPROXY_FALLBACK_IMAGE_PATH")
	IMGPROXY_FALLBACK_IMAGE_URL  = env.String("IMGPROXY_FALLBACK_IMAGE_URL")
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

	IMGPROXY_WATERMARK_DATA.Parse(&c.Base64Data)
	IMGPROXY_WATERMARK_PATH.Parse(&c.Path)
	IMGPROXY_WATERMARK_URL.Parse(&c.URL)

	return c, nil
}

// LoadFallbackStaticConfigFromEnv loads the fallback configuration from the environment
func LoadFallbackStaticConfigFromEnv(c *StaticConfig) (*StaticConfig, error) {
	c = ensure.Ensure(c, NewDefaultStaticConfig)

	IMGPROXY_FALLBACK_IMAGE_DATA.Parse(&c.Base64Data)
	IMGPROXY_FALLBACK_IMAGE_PATH.Parse(&c.Path)
	IMGPROXY_FALLBACK_IMAGE_URL.Parse(&c.URL)

	return c, nil
}
