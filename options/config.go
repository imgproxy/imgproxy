package options

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// URLReplacement represents a URL replacement configuration
type URLReplacement = config.URLReplacement

// Config represents the configuration for options processing
type Config struct {
	// Processing behavior defaults
	StripMetadata     bool // Whether to strip metadata by default
	KeepCopyright     bool // Whether to keep copyright information when stripping metadata
	StripColorProfile bool // Whether to strip color profile by default
	AutoRotate        bool // Whether to auto-rotate images by default
	EnforceThumbnail  bool // Whether to enforce thumbnail extraction by default
	ReturnAttachment  bool // Whether to return images as attachments by default

	// Image processing formats
	SkipProcessingFormats []imagetype.Type // List of formats to skip processing for

	// Presets configuration
	Presets     []string // Available presets
	OnlyPresets bool     // Whether to allow only presets

	// Quality settings
	Quality       int                    // Default quality for image processing
	FormatQuality map[imagetype.Type]int // Quality settings per image format

	// Security and validation
	AllowedProcessingOptions []string // List of allowed processing options

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

	AllowSecurityOptions bool // Whether to allow security options in URLs
}

// NewDefaultConfig creates a new default configuration for options processing
func NewDefaultConfig() Config {
	return Config{
		// Processing behavior defaults (copied from global config defaults)
		StripMetadata:     true,
		KeepCopyright:     true,
		StripColorProfile: true,
		AutoRotate:        true,
		EnforceThumbnail:  false,
		ReturnAttachment:  false,

		OnlyPresets: false,

		// Quality settings (copied from global config defaults)
		Quality: 80,
		FormatQuality: map[imagetype.Type]int{
			imagetype.WEBP: 79,
			imagetype.AVIF: 63,
			imagetype.JXL:  77,
		},

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

		AllowSecurityOptions: false,
	}
}

// LoadConfigFromEnv loads configuration from global config variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	// Copy from global config variables
	c.StripMetadata = config.StripMetadata
	c.KeepCopyright = config.KeepCopyright
	c.StripColorProfile = config.StripColorProfile
	c.AutoRotate = config.AutoRotate
	c.EnforceThumbnail = config.EnforceThumbnail
	c.ReturnAttachment = config.ReturnAttachment

	// Image processing formats
	c.SkipProcessingFormats = slices.Clone(config.SkipProcessingFormats)

	// Presets configuration
	c.Presets = slices.Clone(config.Presets)
	c.OnlyPresets = config.OnlyPresets

	// Quality settings
	c.Quality = config.Quality
	c.FormatQuality = maps.Clone(config.FormatQuality)

	// Security and validation
	c.AllowedProcessingOptions = slices.Clone(config.AllowedProcessingOptions)

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

	c.AllowSecurityOptions = config.AllowSecurityOptions

	return c, nil
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	// Quality validation (copied from global config validation)
	if c.Quality <= 0 {
		return fmt.Errorf("quality should be greater than 0, now - %d", c.Quality)
	} else if c.Quality > 100 {
		return fmt.Errorf("quality can't be greater than 100, now - %d", c.Quality)
	}

	// Format quality validation
	for format, quality := range c.FormatQuality {
		if quality <= 0 {
			return fmt.Errorf("format quality for %s should be greater than 0, now - %d", format, quality)
		} else if quality > 100 {
			return fmt.Errorf("format quality for %s can't be greater than 100, now - %d", format, quality)
		}
	}

	// Arguments separator validation
	if c.ArgumentsSeparator == "" {
		return errors.New("arguments separator cannot be empty")
	}

	return nil
}
