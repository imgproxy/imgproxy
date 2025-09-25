package optionsparser

import (
	"errors"
	"slices"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
)

// URLReplacement represents a URL replacement configuration
type URLReplacement = config.URLReplacement

// Config represents the configuration for options processing
type Config struct {
	// Presets configuration
	Presets     []string // Available presets
	OnlyPresets bool     // Whether to allow only presets

	// Security and validation
	AllowedProcessingOptions []string // List of allowed processing options
	AllowSecurityOptions     bool     // Whether to allow security options in URLs

	// Format preference and enforcement
	AutoWebp    bool // Whether to automatically serve WebP when supported
	EnforceWebp bool // Whether to enforce WebP format
	AutoAvif    bool // Whether to automatically serve AVIF when supported
	EnforceAvif bool // Whether to enforce AVIF format
	AutoJxl     bool // Whether to automatically serve JXL when supported
	EnforceJxl  bool // Whether to enforce JXL format

	// Client hints
	EnableClientHints bool // Whether to enable client hints support

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

		// Format preference and enforcement (copied from global config defaults)
		AutoWebp:    false,
		EnforceWebp: false,
		AutoAvif:    false,
		EnforceAvif: false,
		AutoJxl:     false,
		EnforceJxl:  false,

		// Client hints
		EnableClientHints: false,

		// URL processing (copied from global config defaults)
		ArgumentsSeparator:        ":",
		BaseURL:                   "",
		Base64URLIncludesFilename: false,
	}
}

// LoadConfigFromEnv loads configuration from global config variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	// Presets configuration
	c.Presets = slices.Clone(config.Presets)
	c.OnlyPresets = config.OnlyPresets

	// Security and validation
	c.AllowedProcessingOptions = slices.Clone(config.AllowedProcessingOptions)
	c.AllowSecurityOptions = config.AllowSecurityOptions

	// Format preference and enforcement
	c.AutoWebp = config.AutoWebp
	c.EnforceWebp = config.EnforceWebp
	c.AutoAvif = config.AutoAvif
	c.EnforceAvif = config.EnforceAvif
	c.AutoJxl = config.AutoJxl
	c.EnforceJxl = config.EnforceJxl

	// Client hints
	c.EnableClientHints = config.EnableClientHints

	// URL processing
	c.ArgumentsSeparator = config.ArgumentsSeparator
	c.BaseURL = config.BaseURL
	c.URLReplacements = slices.Clone(config.URLReplacements)
	c.Base64URLIncludesFilename = config.Base64URLIncludesFilename

	return c, nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Arguments separator validation
	if c.ArgumentsSeparator == "" {
		return errors.New("arguments separator cannot be empty")
	}

	return nil
}
