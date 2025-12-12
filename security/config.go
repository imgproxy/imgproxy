package security

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_ALLOW_SECURITY_OPTIONS = env.Bool("IMGPROXY_ALLOW_SECURITY_OPTIONS")
	IMGPROXY_ALLOWED_SOURCES        = env.URLPatterns("IMGPROXY_ALLOWED_SOURCES")
	IMGPROXY_KEYS                   = env.HexSlice("IMGPROXY_KEYS")
	IMGPROXY_SALTS                  = env.HexSlice("IMGPROXY_SALTS")
	IMGPROXY_SIGNATURE_SIZE         = env.Int("IMGPROXY_SIGNATURE_SIZE")
	IMGPROXY_TRUSTED_SIGNATURES     = env.StringSlice("IMGPROXY_TRUSTED_SIGNATURES")

	IMGPROXY_MAX_SRC_RESOLUTION             = env.MegaInt("IMGPROXY_MAX_SRC_RESOLUTION")
	IMGPROXY_MAX_SRC_FILE_SIZE              = env.Int("IMGPROXY_MAX_SRC_FILE_SIZE")
	IMGPROXY_MAX_ANIMATION_FRAMES           = env.Int("IMGPROXY_MAX_ANIMATION_FRAMES")
	IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION = env.MegaInt("IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION")
	IMGPROXY_MAX_RESULT_DIMENSION           = env.Int("IMGPROXY_MAX_RESULT_DIMENSION")
)

// Config is the package-local configuration
type Config struct {
	AllowSecurityOptions bool             // Whether to allow security-related processing options in URLs
	AllowedSources       []*regexp.Regexp // List of allowed source URL patterns (empty = allow all)
	Keys                 [][]byte         // List of the HMAC keys
	Salts                [][]byte         // List of the HMAC salts
	SignatureSize        int              // Size of the HMAC signature in bytes
	TrustedSignatures    []string         // List of trusted signature sources

	MaxSrcResolution            int // Maximum allowed source image resolution
	MaxSrcFileSize              int // Maximum allowed source image file size in bytes
	MaxAnimationFrames          int // Maximum allowed animation frames
	MaxAnimationFrameResolution int // Maximum allowed resolution per animation frame
	MaxResultDimension          int // Maximum allowed result image dimension (width or height)
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		AllowSecurityOptions: false,
		SignatureSize:        32,

		MaxSrcResolution:            50_000_000,
		MaxSrcFileSize:              0,
		MaxAnimationFrames:          1,
		MaxAnimationFrameResolution: 0,
		MaxResultDimension:          0,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_ALLOW_SECURITY_OPTIONS.Parse(&c.AllowSecurityOptions),
		IMGPROXY_ALLOWED_SOURCES.Parse(&c.AllowedSources),
		IMGPROXY_SIGNATURE_SIZE.Parse(&c.SignatureSize),
		IMGPROXY_TRUSTED_SIGNATURES.Parse(&c.TrustedSignatures),

		IMGPROXY_MAX_SRC_RESOLUTION.Parse(&c.MaxSrcResolution),
		IMGPROXY_MAX_SRC_FILE_SIZE.Parse(&c.MaxSrcFileSize),
		IMGPROXY_MAX_ANIMATION_FRAMES.Parse(&c.MaxAnimationFrames),
		IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION.Parse(&c.MaxAnimationFrameResolution),
		IMGPROXY_MAX_RESULT_DIMENSION.Parse(&c.MaxResultDimension),

		IMGPROXY_KEYS.Parse(&c.Keys),
		IMGPROXY_SALTS.Parse(&c.Salts),
	)

	return c, err
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxSrcResolution <= 0 {
		return IMGPROXY_MAX_SRC_RESOLUTION.ErrorZeroOrNegative()
	}

	if c.MaxSrcFileSize < 0 {
		return IMGPROXY_MAX_SRC_FILE_SIZE.ErrorNegative()
	}

	if c.MaxAnimationFrames <= 0 {
		return IMGPROXY_MAX_ANIMATION_FRAMES.ErrorZeroOrNegative()
	}

	if len(c.Keys) != len(c.Salts) {
		return fmt.Errorf(
			"number of keys and number of salts should be equal. Keys: %d, salts: %d",
			len(c.Keys), len(c.Salts),
		)
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
