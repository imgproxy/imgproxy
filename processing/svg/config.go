package svg

import (
	"errors"
	"fmt"

	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/env"
)

var (
	IMGPROXY_SANITIZE_SVG       = env.Bool("IMGPROXY_SANITIZE_SVG")
	IMGPROXY_SANITIZE_SVG_TAGS  = env.StringSlice("IMGPROXY_SANITIZE_SVG_TAGS")
	IMGPROXY_SANITIZE_SVG_ATTRS = env.StringSlice("IMGPROXY_SANITIZE_SVG_ATTRS")
)

// Config holds SVG-specific configuration
type Config struct {
	Sanitize bool // Sanitize SVG content for security

	// Tags and Attrs adjust the default allowlists of SVG elements and
	// unprefixed attributes. Each entry is "+name" (add), "-name" (remove) or
	// "name" (replace the defaults with the bare entries). See applyListOverrides.
	Tags  []string
	Attrs []string
}

// NewDefaultConfig creates a new Config instance with default values
func NewDefaultConfig() Config {
	return Config{
		Sanitize: true, // By default, sanitize SVG for security
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_SANITIZE_SVG.Parse(&c.Sanitize),
		IMGPROXY_SANITIZE_SVG_TAGS.Parse(&c.Tags),
		IMGPROXY_SANITIZE_SVG_ATTRS.Parse(&c.Attrs),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	_, tagsErr := applyListOverrides(defaultTags, c.Tags)
	if tagsErr != nil {
		tagsErr = fmt.Errorf("%s: %w", IMGPROXY_SANITIZE_SVG_TAGS.Name, tagsErr)
	}

	_, attrsErr := applyListOverrides(defaultAttrs, c.Attrs)
	if attrsErr != nil {
		attrsErr = fmt.Errorf("%s: %w", IMGPROXY_SANITIZE_SVG_ATTRS.Name, attrsErr)
	}

	return errors.Join(tagsErr, attrsErr)
}
