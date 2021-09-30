package security

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var ErrSourceResolutionTooBig = ierrors.New(422, "Source image resolution is too big", "Invalid source image")

func CheckDimensions(width, height int) error {
	if width*height > config.MaxSrcResolution {
		return ErrSourceResolutionTooBig
	}

	return nil
}
