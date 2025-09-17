package processing

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

// Config holds processing-related configuration.
type Config struct {
	PreferredFormats    []imagetype.Type
	WatermarkOpacity    float64
	DisableShrinkOnLoad bool
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

	c.PreferredFormats = config.PreferredFormats

	return c, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	filtered := c.PreferredFormats[:0]

	for _, t := range c.PreferredFormats {
		if !vips.SupportsSave(t) {
			log.Warnf("%s can't be a preferred format as it's saving is not supported", t)
		} else {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return errors.New("no supported preferred formats specified")
	}

	c.PreferredFormats = filtered

	if c.WatermarkOpacity <= 0 {
		return errors.New("watermark opacity should be greater than 0")
	} else if c.WatermarkOpacity > 1 {
		return errors.New("watermark opacity should be less than or equal to 1")
	}

	return nil
}
