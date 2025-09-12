package security

import (
	"github.com/imgproxy/imgproxy/v3/config"
)

type Options struct {
	MaxSrcResolution            int
	MaxSrcFileSize              int
	MaxAnimationFrames          int
	MaxAnimationFrameResolution int
	MaxResultDimension          int
}

// NOTE: Remove this function in imgproxy v4
// TODO: Replace this with security.NewOptions() when ProcessingOptions gets config
func DefaultOptions() Options {
	return Options{
		MaxSrcResolution:            config.MaxSrcResolution,
		MaxSrcFileSize:              config.MaxSrcFileSize,
		MaxAnimationFrames:          config.MaxAnimationFrames,
		MaxAnimationFrameResolution: config.MaxAnimationFrameResolution,
		MaxResultDimension:          config.MaxResultDimension,
	}
}

func IsSecurityOptionsAllowed() error {
	if config.AllowSecurityOptions {
		return nil
	}

	return newSecurityOptionsError()
}

// CheckDimensions checks if the given dimensions are within the allowed limits
func (o *Options) CheckDimensions(width, height, frames int) error {
	frames = max(frames, 1)

	if frames > 1 && o.MaxAnimationFrameResolution > 0 {
		if width*height > o.MaxAnimationFrameResolution {
			return newImageResolutionError("Source image frame resolution is too big")
		}
	} else {
		if width*height*frames > o.MaxSrcResolution {
			return newImageResolutionError("Source image resolution is too big")
		}
	}

	return nil
}
