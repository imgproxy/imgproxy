package processing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

var (
	preferredFormats = env.Define(
		"IMGPROXY_PREFERRED_FORMATS",
		"Preferred image formats, comma-separated",
		"jpeg, png, webp, tiff, avif, heic, gif, jp2", // what else?
		parsePreferredFormats,
		[]imagetype.Type{imagetype.JPEG, imagetype.PNG, imagetype.GIF},
	)
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
		PreferredFormats: preferredFormats.Default(),
	}
}

// NewConfig creates a new Config instance with the given parameters.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.WatermarkOpacity = config.WatermarkOpacity
	c.DisableShrinkOnLoad = config.DisableShrinkOnLoad
	c.UseLinearColorspace = config.UseLinearColorspace

	if err := preferredFormats.Get(&c.PreferredFormats); err != nil {
		return nil, err
	}

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
			log.Warnf("%s can't be a preferred format as it's saving is not supported", t)
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

func parsePreferredFormats(s string) ([]imagetype.Type, error) {
	parts := strings.Split(s, ",")
	it := make([]imagetype.Type, 0, len(parts))

	for _, p := range parts {
		part := strings.TrimSpace(p)

		// For every part passed through the environment variable,
		// check if it matches any of the image types defined in
		// the imagetype package or return error.
		t, ok := imagetype.GetTypeByName(part)
		if !ok {
			return nil, fmt.Errorf("unknown image format: %s", part)
		}
		it = append(it, t)
	}

	return it, nil
}
