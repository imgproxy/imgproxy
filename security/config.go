package security

import (
	"fmt"
	"regexp"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	log "github.com/sirupsen/logrus"
)

// OptionsConfig represents the configuration for processing limits and security options
type OptionsConfig struct {
	MaxSrcResolution            int // Maximum allowed source image resolution (width × height)
	MaxSrcFileSize              int // Maximum allowed source file size in bytes (0 = unlimited)
	MaxAnimationFrames          int // Maximum number of frames allowed in animated images
	MaxAnimationFrameResolution int // Maximum resolution allowed for each frame in animated images (0 = unlimited)
	MaxResultDimension          int // Maximum allowed dimension (width or height) for result images (0 = unlimited)
}

// NewDefaultOptionsConfig returns a new OptionsConfig instance with default values
func NewDefaultOptionsConfig() OptionsConfig {
	return OptionsConfig{
		MaxSrcResolution:            50000000,
		MaxSrcFileSize:              0,
		MaxAnimationFrames:          1,
		MaxAnimationFrameResolution: 0,
		MaxResultDimension:          0,
	}
}

// LoadOptionsConfigFromEnv loads OptionsConfig from global config variables
func LoadOptionsConfigFromEnv(c *OptionsConfig) (*OptionsConfig, error) {
	c.MaxSrcResolution = config.MaxSrcResolution
	c.MaxSrcFileSize = config.MaxSrcFileSize
	c.MaxAnimationFrames = config.MaxAnimationFrames
	c.MaxAnimationFrameResolution = config.MaxAnimationFrameResolution
	c.MaxResultDimension = config.MaxResultDimension

	return c, nil
}

// Validate validates the OptionsConfig values
func (c *OptionsConfig) Validate() error {
	if c.MaxSrcResolution <= 0 {
		return fmt.Errorf("max src resolution should be greater than 0, now - %d", c.MaxSrcResolution)
	}

	if c.MaxSrcFileSize < 0 {
		return fmt.Errorf("max src file size should be greater than or equal to 0, now - %d", c.MaxSrcFileSize)
	}

	if c.MaxAnimationFrames <= 0 {
		return fmt.Errorf("max animation frames should be greater than 0, now - %d", c.MaxAnimationFrames)
	}

	return nil
}

// Config is the package-local configuration
type Config struct {
	Options              OptionsConfig    // Processing limits and security options
	AllowSecurityOptions bool             // Whether to allow security-related processing options in URLs
	AllowedSources       []*regexp.Regexp // List of allowed source URL patterns (empty = allow all)
	Keys                 [][]byte         // List of the HMAC keys
	Salts                [][]byte         // List of the HMAC salts
	SignatureSize        int              // Size of the HMAC signature in bytes
	TrustedSignatures    []string         // List of trusted signature sources
}

// NewDefaultConfig returns a new Config instance with default values.
func NewDefaultConfig() Config {
	return Config{
		Options:              NewDefaultOptionsConfig(),
		AllowSecurityOptions: false,
		SignatureSize:        32,
	}
}

// LoadConfigFromEnv overrides configuration variables from environment
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	if _, err := LoadOptionsConfigFromEnv(&c.Options); err != nil {
		return nil, err
	}

	c.AllowSecurityOptions = config.AllowSecurityOptions
	c.AllowedSources = config.AllowedSources
	c.Keys = config.Keys
	c.Salts = config.Salts
	c.SignatureSize = config.SignatureSize
	c.TrustedSignatures = config.TrustedSignatures

	return c, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Options.Validate(); err != nil {
		return err
	}

	if len(c.Keys) != len(c.Salts) {
		return fmt.Errorf("number of keys and number of salts should be equal. Keys: %d, salts: %d", len(c.Keys), len(c.Salts))
	}

	if len(c.Keys) == 0 {
		log.Warning("No keys defined, so signature checking is disabled")
	}

	if len(c.Salts) == 0 {
		log.Warning("No salts defined, so signature checking is disabled")
	}

	if c.SignatureSize < 1 || c.SignatureSize > 32 {
		return fmt.Errorf("signature size should be within 1 and 32, now - %d", c.SignatureSize)
	}

	return nil
}
