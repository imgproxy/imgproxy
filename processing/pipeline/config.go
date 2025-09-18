package pipeline

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// Config holds pipeline-related configuration.
type Config struct {
	WatermarkOpacity    float64
	DisableShrinkOnLoad bool
	UseLinearColorspace bool
}

// NewConfig creates a new Config instance with the given parameters.
func NewDefaultConfig() Config {
	return Config{
		WatermarkOpacity: 1,
	}
}

// NewConfig creates a new Config instance with the given parameters.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.WatermarkOpacity = config.WatermarkOpacity
	c.DisableShrinkOnLoad = config.DisableShrinkOnLoad
	c.UseLinearColorspace = config.UseLinearColorspace

	return c, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.WatermarkOpacity <= 0 {
		return errors.New("watermark opacity should be greater than 0")
	} else if c.WatermarkOpacity > 1 {
		return errors.New("watermark opacity should be less than or equal to 1")
	}

	return nil
}
