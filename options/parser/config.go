package optionsparser

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

type URLReplacement = env.URLReplacement

var (
	IMGPROXY_PRESETS_SEPARATOR            = env.String("IMGPROXY_PRESETS_SEPARATOR")
	IMGPROXY_PRESETS                      = env.StringSliceSep("IMGPROXY_PRESETS", IMGPROXY_PRESETS_SEPARATOR)
	IMGPROXY_PRESETS_PATH                 = env.StringSliceFile("IMGPROXY_PRESETS_PATH")
	IMGPROXY_ONLY_PRESETS                 = env.Bool("IMGPROXY_ONLY_PRESETS")
	IMGPROXY_ALLOWED_PROCESSING_OPTIONS   = env.StringSlice("IMGPROXY_ALLOWED_PROCESSING_OPTIONS")
	IMGPROXY_ALLOW_SECURITY_OPTIONS       = env.Bool("IMGPROXY_ALLOW_SECURITY_OPTIONS")
	IMGPROXY_ARGUMENTS_SEPARATOR          = env.String("IMGPROXY_ARGUMENTS_SEPARATOR")
	IMGPROXY_BASE_URL                     = env.String("IMGPROXY_BASE_URL")
	IMGPROXY_URL_REPLACEMENTS             = env.URLReplacements("IMGPROXY_URL_REPLACEMENTS")
	IMGPROXY_BASE64_URL_INCLUDES_FILENAME = env.Bool("IMGPROXY_BASE64_URL_INCLUDES_FILENAME")
)

// Config represents the configuration for options processing
type Config struct {
	// Presets configuration
	Presets     []string // Available presets
	OnlyPresets bool     // Whether to allow only presets

	// Security and validation
	AllowedProcessingOptions []string // List of allowed processing options
	AllowSecurityOptions     bool     // Whether to allow security options in URLs

	// URL processing
	ArgumentsSeparator        string           // Separator for URL arguments
	BaseURL                   string           // Base URL for relative URLs
	URLReplacements           []URLReplacement // URL replacement rules
	Base64URLIncludesFilename bool             // Whether base64 URLs include filename
}

// NewDefaultConfig creates a new default configuration for options processing
func NewDefaultConfig() Config {
	return Config{
		// Presets configuration
		OnlyPresets: false,

		// Security and validation
		AllowSecurityOptions: false,

		// URL processing (copied from global config defaults)
		ArgumentsSeparator:        ":",
		BaseURL:                   "",
		Base64URLIncludesFilename: false,
	}
}

// LoadConfigFromEnv loads configuration from global config variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	c.Presets = make([]string, 0)
	c.URLReplacements = make([]URLReplacement, 0)

	var presetsFromFile []string

	err := errors.Join(
		IMGPROXY_ONLY_PRESETS.Parse(&c.OnlyPresets),
		IMGPROXY_PRESETS.Parse(&c.Presets),
		IMGPROXY_PRESETS_PATH.Parse(&presetsFromFile),

		// Security and validation
		IMGPROXY_ALLOWED_PROCESSING_OPTIONS.Parse(&c.AllowedProcessingOptions),
		IMGPROXY_ALLOW_SECURITY_OPTIONS.Parse(&c.AllowSecurityOptions),

		// URL processing
		IMGPROXY_ARGUMENTS_SEPARATOR.Parse(&c.ArgumentsSeparator),
		IMGPROXY_BASE_URL.Parse(&c.BaseURL),
		IMGPROXY_BASE64_URL_INCLUDES_FILENAME.Parse(&c.Base64URLIncludesFilename),

		IMGPROXY_URL_REPLACEMENTS.Parse(&c.URLReplacements),
	)

	c.Presets = append(c.Presets, presetsFromFile...)

	return c, err
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Arguments separator validation
	if c.ArgumentsSeparator == "" {
		return IMGPROXY_ARGUMENTS_SEPARATOR.ErrorEmpty()
	}

	return nil
}
