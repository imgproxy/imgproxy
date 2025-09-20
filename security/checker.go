package security

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

// Checker represents the security package instance
type Checker struct {
	config *Config
}

// New creates a new Security instance
func New(config *Config) (*Checker, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Checker{
		config: config,
	}, nil
}

// NewOptions creates a new [security.Options] instance
// filling it from [options.Options].
// If opts is nil, it returns default [security.Options].
func (s *Checker) NewOptions(opts options.Options) (secops Options) {
	secops = s.config.DefaultOptions
	if opts == nil {
		return
	}

	secops.MaxSrcResolution = options.GetInt(
		opts, keys.MaxSrcResolution, secops.MaxSrcResolution,
	)
	secops.MaxSrcFileSize = options.GetInt(
		opts, keys.MaxSrcFileSize, secops.MaxSrcFileSize,
	)
	secops.MaxAnimationFrames = options.GetInt(
		opts, keys.MaxAnimationFrames, secops.MaxAnimationFrames,
	)
	secops.MaxAnimationFrameResolution = options.GetInt(
		opts, keys.MaxAnimationFrameResolution, secops.MaxAnimationFrameResolution,
	)
	secops.MaxResultDimension = options.GetInt(
		opts, keys.MaxResultDimension, secops.MaxResultDimension,
	)

	return
}
