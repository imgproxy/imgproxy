package optionsparser

import (
	"errors"
	"os"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

type URLReplacement = env.URLReplacement

const (
	PresetsFlagName = "presets" // --presets flag name
)

var (
	IMGPROXY_PRESETS_SEPARATOR            = env.Describe("IMGPROXY_PRESETS_SEPARATOR", "string")
	IMGPROXY_PRESETS                      = env.Describe("IMGPROXY_PRESETS", "separated list of strings (see IMGPROXY_PRESETS_SEPARATOR)") //nolint:lll
	IMGPROXY_ONLY_PRESETS                 = env.Describe("IMGPROXY_ONLY_PRESETS", "boolean")
	IMGPROXY_ALLOWED_PROCESSING_OPTIONS   = env.Describe("IMGPROXY_ALLOWED_PROCESSING_OPTIONS", "comma-separated list of strings") //nolint:lll
	IMGPROXY_ALLOW_SECURITY_OPTIONS       = env.Describe("IMGPROXY_ALLOW_SECURITY_OPTIONS", "boolean")
	IMGPROXY_ARGUMENTS_SEPARATOR          = env.Describe("IMGPROXY_ARGUMENTS_SEPARATOR", "string")
	IMGPROXY_BASE_URL                     = env.Describe("IMGPROXY_BASE_URL", "string")
	IMGPROXY_URL_REPLACEMENTS             = env.Describe("IMGPROXY_URL_REPLACEMENTS", "comma-separated list of key=value pairs") //nolint:lll
	IMGPROXY_BASE64_URL_INCLUDES_FILENAME = env.Describe("IMGPROXY_BASE64_URL_INCLUDES_FILENAME", "boolean")

	// PRESETS_PATH Artificial env.desc for --presets flag
	PRESETS_PATH = env.Describe("--"+PresetsFlagName, "path to presets file")
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

	sep := ","
	var presetsPath string

	// NOTE: Here is the workaround for reading --presets flag from CLI.
	// Otherwise, we'd have to either store it in a global variable or
	// pass cli.Command down the call stack.
	for i, arg := range os.Args {
		if arg == "--"+PresetsFlagName && i+1 < len(os.Args) {
			presetsPath = os.Args[i+1]
			break
		}
	}

	c.Presets = make([]string, 0)
	c.URLReplacements = make([]URLReplacement, 0)

	err := errors.Join(
		env.String(&sep, IMGPROXY_PRESETS_SEPARATOR),
		env.StringSliceSep(&c.Presets, IMGPROXY_PRESETS, sep),
		env.Bool(&c.OnlyPresets, IMGPROXY_ONLY_PRESETS),

		// Security and validation
		env.StringSlice(&c.AllowedProcessingOptions, IMGPROXY_ALLOWED_PROCESSING_OPTIONS),
		env.Bool(&c.AllowSecurityOptions, IMGPROXY_ALLOW_SECURITY_OPTIONS),

		// URL processing
		env.String(&c.ArgumentsSeparator, IMGPROXY_ARGUMENTS_SEPARATOR),
		env.String(&c.BaseURL, IMGPROXY_BASE_URL),
		env.Bool(&c.Base64URLIncludesFilename, IMGPROXY_BASE64_URL_INCLUDES_FILENAME),

		env.StringSliceFile(&c.Presets, PRESETS_PATH, presetsPath),
		env.URLReplacements(&c.URLReplacements, IMGPROXY_URL_REPLACEMENTS),
	)

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
