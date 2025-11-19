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

// MaxSrcResolution returns the maximum allowed source image resolution
func (s *Checker) MaxSrcResolution(o *options.Options) int {
	return o.GetInt(keys.MaxSrcResolution, s.config.MaxSrcResolution)
}

// MaxSrcFileSize returns the maximum allowed source file size
func (s *Checker) MaxSrcFileSize(o *options.Options) int {
	return o.GetInt(keys.MaxSrcFileSize, s.config.MaxSrcFileSize)
}

// MaxAnimationFrames returns the maximum allowed animation frames
func (s *Checker) MaxAnimationFrames(o *options.Options) int {
	return o.GetInt(keys.MaxAnimationFrames, s.config.MaxAnimationFrames)
}

// MaxAnimationFrameResolution returns the maximum allowed animation frame resolution
func (s *Checker) MaxAnimationFrameResolution(o *options.Options) int {
	return o.GetInt(
		keys.MaxAnimationFrameResolution,
		s.config.MaxAnimationFrameResolution,
	)
}

// MaxResultDimension returns the maximum allowed result image dimension
func (s *Checker) MaxResultDimension(o *options.Options) int {
	return o.GetInt(keys.MaxResultDimension, s.config.MaxResultDimension)
}

// CheckDimensions checks if the given dimensions are within the allowed limits
func (s *Checker) CheckDimensions(o *options.Options, width, height, frames int) error {
	frames = max(frames, 1)

	maxFrameRes := s.MaxAnimationFrameResolution(o)

	if frames > 1 && maxFrameRes > 0 {
		if width*height > maxFrameRes {
			return newImageResolutionError("Source image frame resolution is too big")
		}
		return nil
	}

	if width*height*frames > s.MaxSrcResolution(o) {
		return newImageResolutionError("Source image resolution is too big")
	}

	return nil
}
