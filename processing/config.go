package processing

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Config holds pipeline-related configuration.
type Config struct {
	PreferredFormats    []imagetype.Type
	WatermarkOpacity    float64
	DisableShrinkOnLoad bool
	UseLinearColorspace bool
}

// NewConfig creates a new Config instance with the given parameters.
func NewDefaultConfig() Config {
	return Config{
		WatermarkOpacity: 1,
		PreferredFormats: []imagetype.Type{
			imagetype.JPEG,
			imagetype.PNG,
			imagetype.GIF,
		},
	}
}

// NewConfig creates a new Config instance with the given parameters.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.WatermarkOpacity = config.WatermarkOpacity
	c.DisableShrinkOnLoad = config.DisableShrinkOnLoad
	c.UseLinearColorspace = config.UseLinearColorspace
	c.PreferredFormats = config.PreferredFormats

	return c, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.WatermarkOpacity <= 0 {
		return errors.New("watermark opacity should be greater than 0")
	} else if c.WatermarkOpacity > 1 {
		return errors.New("watermark opacity should be less than or equal to 1")
	}

	filtered := c.PreferredFormats[:0]

	for _, t := range c.PreferredFormats {
		if !vips.SupportsSave(t) {
			slog.Warn(fmt.Sprintf("%s can't be a preferred format as it's saving is not supported", t))
		} else {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return errors.New("no supported preferred formats specified")
	}

	c.PreferredFormats = filtered

	return nil
}
