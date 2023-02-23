package security

import (
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imath"
)

var ErrSourceResolutionTooBig = ierrors.New(422, "Source image resolution is too big", "Invalid source image")
var ErrSourceFrameResolutionTooBig = ierrors.New(422, "Source image frame resolution is too big", "Invalid source image")

func CheckDimensions(width, height, frames int, opts Options) error {
	frames = imath.Max(frames, 1)

	if frames > 1 && opts.MaxAnimationFrameResolution > 0 {
		if width*height > opts.MaxAnimationFrameResolution {
			return ErrSourceFrameResolutionTooBig
		}
	} else {
		if width*height*frames > opts.MaxSrcResolution {
			return ErrSourceResolutionTooBig
		}
	}

	return nil
}
