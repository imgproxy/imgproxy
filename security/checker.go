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
func (s *Checker) NewOptions(opts *options.Options) (secops Options) {
	secops = s.config.DefaultOptions
	if opts == nil {
		return
	}

	secops.MaxSrcResolution = opts.GetInt(
		keys.MaxSrcResolution, secops.MaxSrcResolution,
	)
	secops.MaxSrcFileSize = opts.GetInt(
		keys.MaxSrcFileSize, secops.MaxSrcFileSize,
	)
	secops.MaxAnimationFrames = opts.GetInt(
		keys.MaxAnimationFrames, secops.MaxAnimationFrames,
	)
	secops.MaxAnimationFrameResolution = opts.GetInt(
		keys.MaxAnimationFrameResolution, secops.MaxAnimationFrameResolution,
	)
	secops.MaxResultDimension = opts.GetInt(
		keys.MaxResultDimension, secops.MaxResultDimension,
	)

	return
}
