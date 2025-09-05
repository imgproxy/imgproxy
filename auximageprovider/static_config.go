package auximageprovider

import "github.com/imgproxy/imgproxy/v3/config"

// StaticConfig holds the configuration for the auxiliary image provider
type StaticConfig struct {
	Base64Data string
	Path       string
	URL        string
}

// NewDefaultStaticConfig creates a new default configuration for the auxiliary image provider
func NewDefaultStaticConfig() *StaticConfig {
	return &StaticConfig{
		Base64Data: "",
		Path:       "",
		URL:        "",
	}
}

// LoadWatermarkStaticConfigFromEnv loads the watermark configuration from the environment
func LoadWatermarkStaticConfigFromEnv(c *StaticConfig) (*StaticConfig, error) {
	if c == nil {
		c = NewDefaultStaticConfig()
	}

	c.Base64Data = config.WatermarkData
	c.Path = config.WatermarkPath
	c.URL = config.WatermarkURL

	return c, nil
}

// LoadFallbackStaticConfigFromEnv loads the fallback configuration from the environment
func LoadFallbackStaticConfigFromEnv(c *StaticConfig) (*StaticConfig, error) {
	if c == nil {
		c = NewDefaultStaticConfig()
	}

	c.Base64Data = config.FallbackImageData
	c.Path = config.FallbackImagePath
	c.URL = config.FallbackImageURL

	return c, nil
}
