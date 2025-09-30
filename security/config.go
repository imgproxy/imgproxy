package security

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_ALLOW_SECURITY_OPTIONS = env.Describe("IMGPROXY_ALLOW_SECURITY_OPTIONS", "boolean")
	IMGPROXY_ALLOWED_SOURCES        = env.Describe("IMGPROXY_ALLOWED_SOURCES", "comma-separated lists of regexes")
	IMGPROXY_KEYS                   = env.Describe("IMGPROXY_KEYS", "comma-separated list of hex strings")
	IMGPROXY_SALTS                  = env.Describe("IMGPROXY_SALTS", "comma-separated list of hex strings")
	IMGPROXY_SIGNATURE_SIZE         = env.Describe("IMGPROXY_SIGNATURE_SIZE", "number between 1 and 32")
	IMGPROXY_TRUSTED_SIGNATURES     = env.Describe("IMGPROXY_TRUSTED_SIGNATURES", "comma-separated list of strings")

	IMGPROXY_MAX_SRC_RESOLUTION             = env.Describe("IMGPROXY_MAX_SRC_RESOLUTION", "number > 0")
	IMGPROXY_MAX_SRC_FILE_SIZE              = env.Describe("IMGPROXY_MAX_SRC_FILE_SIZE", "number >= 0")
	IMGPROXY_MAX_ANIMATION_FRAMES           = env.Describe("IMGPROXY_MAX_ANIMATION_FRAMES", "number > 0")
	IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION = env.Describe("IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION", "number > 0")
	IMGPROXY_MAX_RESULT_DIMENSION           = env.Describe("IMGPROXY_MAX_RESULT_DIMENSION", "number > 0")
)

// Config is the package-local configuration
type Config struct {
	AllowSecurityOptions bool             // Whether to allow security-related processing options in URLs
	AllowedSources       []*regexp.Regexp // List of allowed source URL patterns (empty = allow all)
	Keys                 [][]byte         // List of the HMAC keys
	Salts                [][]byte         // List of the HMAC salts
	SignatureSize        int              // Size of the HMAC signature in bytes
	TrustedSignatures    []string         // List of trusted signature sources
	DefaultOptions       Options          // Default security options
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		DefaultOptions: Options{
			MaxSrcResolution:            50_000_000,
			MaxSrcFileSize:              0,
			MaxAnimationFrames:          1,
			MaxAnimationFrameResolution: 0,
			MaxResultDimension:          0,
		},
		AllowSecurityOptions: false,
		SignatureSize:        32,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		env.Bool(&c.AllowSecurityOptions, IMGPROXY_ALLOW_SECURITY_OPTIONS),
		env.Patterns(&c.AllowedSources, IMGPROXY_ALLOWED_SOURCES),
		env.Int(&c.SignatureSize, IMGPROXY_SIGNATURE_SIZE),
		env.StringSlice(&c.TrustedSignatures, IMGPROXY_TRUSTED_SIGNATURES),

		env.MegaInt(&c.DefaultOptions.MaxSrcResolution, IMGPROXY_MAX_SRC_RESOLUTION),
		env.Int(&c.DefaultOptions.MaxSrcFileSize, IMGPROXY_MAX_SRC_FILE_SIZE),
		env.Int(&c.DefaultOptions.MaxAnimationFrames, IMGPROXY_MAX_ANIMATION_FRAMES),
		env.MegaInt(&c.DefaultOptions.MaxAnimationFrameResolution, IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION),
		env.Int(&c.DefaultOptions.MaxResultDimension, IMGPROXY_MAX_RESULT_DIMENSION),

		env.HexSlice(&c.Keys, IMGPROXY_KEYS),
		env.HexSlice(&c.Salts, IMGPROXY_SALTS),
	)

	return c, err
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultOptions.MaxSrcResolution <= 0 {
		return IMGPROXY_MAX_SRC_RESOLUTION.ErrorZeroOrNegative()
	}

	if c.DefaultOptions.MaxSrcFileSize < 0 {
		return IMGPROXY_MAX_SRC_FILE_SIZE.ErrorNegative()
	}

	if c.DefaultOptions.MaxAnimationFrames <= 0 {
		return IMGPROXY_MAX_ANIMATION_FRAMES.ErrorZeroOrNegative()
	}

	if len(c.Keys) != len(c.Salts) {
		return fmt.Errorf("number of keys and number of salts should be equal. Keys: %d, salts: %d", len(c.Keys), len(c.Salts))
	}

	if len(c.Keys) == 0 {
		IMGPROXY_KEYS.Warn("No keys defined, signature checking is disabled")
	}

	if len(c.Salts) == 0 {
		IMGPROXY_SALTS.Warn("No salts defined, signature checking is disabled")
	}

	if c.SignatureSize < 1 || c.SignatureSize > 32 {
		return IMGPROXY_SIGNATURE_SIZE.Errorf("invalid size")
	}

	return nil
}
