package security

// Security options (part of processing options)
type Options struct {
	MaxSrcResolution            int
	MaxSrcFileSize              int
	MaxAnimationFrames          int
	MaxAnimationFrameResolution int
	MaxResultDimension          int
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
