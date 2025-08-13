package security

import (
	"github.com/imgproxy/imgproxy/v3/imath"
)

func CheckDimensions(width, height, frames int, opts Options) error {
	frames = imath.Max(frames, 1)

	if frames > 1 && opts.MaxAnimationFrameResolution > 0 {
		if width*height > opts.MaxAnimationFrameResolution {
			return newImageResolutionError("Source image frame resolution is too big")
		}
	} else {
		if width*height*frames > opts.MaxSrcResolution {
			return newImageResolutionError("Source image resolution is too big")
		}
	}

	return nil
}
