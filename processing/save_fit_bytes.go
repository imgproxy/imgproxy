package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// saveImageToFitBytes tries to save the image to fit into the specified max bytes
// by lowering the quality. It returns the image data that fits the requirement
// or the best effort data if it was not possible to fit into the limit.
func saveImageToFitBytes(
	ctx context.Context,
	img *vips.Image,
	format imagetype.Type,
	startQuality int,
	target int,
) (imagedata.ImageData, error) {
	var newQuality int

	// Start with the specified quality and go down from there.
	quality := startQuality

	// We will probably save the image multiple times, so we need to process its pixels
	// to ensure that it is in random access mode.
	if err := img.CopyMemory(); err != nil {
		return nil, err
	}

	for {
		// Check for timeout or cancellation before each attempt as we might spend too much
		// time processing the image or making previous attempts.
		if err := server.CheckTimeout(ctx); err != nil {
			return nil, err
		}

		imgdata, err := img.Save(format, quality)
		if err != nil {
			return nil, err
		}

		size, err := imgdata.Size()
		if err != nil {
			imgdata.Close()
			return nil, err
		}

		// If we fit the limit or quality is too low, return the result.
		if size <= target || quality <= 10 {
			return imgdata, err
		}

		// We don't need the image data anymore, close it to free resources.
		imgdata.Close()

		// Tune quality for the next attempt based on how much we exceed the limit.
		delta := float64(size) / float64(target)
		switch {
		case delta > 3:
			newQuality = imath.Scale(quality, 0.25)
		case delta > 1.5:
			newQuality = imath.Scale(quality, 0.5)
		default:
			newQuality = imath.Scale(quality, 0.75)
		}

		// Ensure that quality is always lowered, even if the scaling
		// doesn't change it due to rounding.
		// Also, ensure that quality doesn't go below the minimum.
		quality = max(1, min(quality-1, newQuality))
	}
}
