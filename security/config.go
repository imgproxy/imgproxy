package security

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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
			MaxSrcResolution:            50000000,
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

	c.AllowSecurityOptions = config.AllowSecurityOptions
	c.AllowedSources = config.AllowedSources
	c.Keys = config.Keys
	c.Salts = config.Salts
	c.SignatureSize = config.SignatureSize
	c.TrustedSignatures = config.TrustedSignatures

	c.DefaultOptions.MaxSrcResolution = config.MaxSrcResolution
	c.DefaultOptions.MaxSrcFileSize = config.MaxSrcFileSize
	c.DefaultOptions.MaxAnimationFrames = config.MaxAnimationFrames
	c.DefaultOptions.MaxAnimationFrameResolution = config.MaxAnimationFrameResolution
	c.DefaultOptions.MaxResultDimension = config.MaxResultDimension

	return c, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultOptions.MaxSrcResolution <= 0 {
		return fmt.Errorf("max src resolution should be greater than 0, now - %d", c.DefaultOptions.MaxSrcResolution)
	}

	if c.DefaultOptions.MaxSrcFileSize < 0 {
		return fmt.Errorf("max src file size should be greater than or equal to 0, now - %d", c.DefaultOptions.MaxSrcFileSize)
	}

	if c.DefaultOptions.MaxAnimationFrames <= 0 {
		return fmt.Errorf("max animation frames should be greater than 0, now - %d", c.DefaultOptions.MaxAnimationFrames)
	}

	if len(c.Keys) != len(c.Salts) {
		return fmt.Errorf("number of keys and number of salts should be equal. Keys: %d, salts: %d", len(c.Keys), len(c.Salts))
	}

	if len(c.Keys) == 0 {
		slog.Warn("No keys defined, so signature checking is disabled")
	}

	if len(c.Salts) == 0 {
		slog.Warn("No salts defined, so signature checking is disabled")
	}

	if c.SignatureSize < 1 || c.SignatureSize > 32 {
		return fmt.Errorf("signature size should be within 1 and 32, now - %d", c.SignatureSize)
	}

	return nil
}
