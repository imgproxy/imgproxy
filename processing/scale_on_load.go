package processing

import (
	"log/slog"
	"math"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func (p *Processor) canScaleOnLoad(c *Context, shrink float64) bool {
	if c.ImgData == nil || shrink == 1 {
		return false
	}

	if c.ImgData.Format().IsVector() {
		return true
	}

	if p.config.DisableShrinkOnLoad || shrink <= 1 {
		return false
	}

	return c.ImgData.Format() == imagetype.JPEG ||
		c.ImgData.Format() == imagetype.WEBP ||
		c.ImgData.Format().SupportsThumbnail()
}

func calcJpegShink(shrink float64) float64 {
	switch {
	case shrink >= 8:
		return 8
	case shrink >= 4:
		return 4
	case shrink >= 2:
		return 2
	}

	return 1
}

func (p *Processor) scaleOnLoad(c *Context) error {
	// Get the preshrink value based on the requested scales.
	// We calculate it based on the image dimensions that we would get
	// with the current scales.
	// We can't just use c.WScale and c.HScale since this may lead to
	// overshrinking when only one target dimension is set.
	wshrink := float64(c.SrcWidth) / float64(imath.Scale(c.SrcWidth, c.WScale))
	hshrink := float64(c.SrcHeight) / float64(imath.Scale(c.SrcHeight, c.HScale))
	preshrink := min(wshrink, hshrink)

	// For vector images, apply the vector base shrink.
	// We might set it in the [Processor.vectorGuardScale] step in case the image
	// is too large.
	if c.ImgData != nil && c.ImgData.Format().IsVector() {
		preshrink *= c.VectorBaseShrink
	}

	// Check if we can and should scale the image on load
	if !p.canScaleOnLoad(c, preshrink) {
		return nil
	}

	// We will load the prescaled image into this new image.
	// On success, we will swap it with the original image in the context,
	// so we can safely clear it on function exit.
	newImg := new(vips.Image)
	defer newImg.Clear()

	loadThumbnail := c.ImgData.Format().SupportsThumbnail()

	if loadThumbnail {
		// If the image supports embedded thumbnails, try to load it
		if err := newImg.LoadThumbnail(c.ImgData); err != nil {
			slog.Debug("Can't load thumbnail: %s", "error", err)
			return nil
		}
	} else {
		// JPEG shrink-on-load must be 1, 2, 4 or 8.
		// We need to normalize it before passing to libvips.
		// For other formats, we can pass any float value.
		if c.ImgData.Format() == imagetype.JPEG {
			preshrink = calcJpegShink(preshrink)
		}

		// if preshrink is 1, we can skip reloading the image
		if preshrink == 1 {
			return nil
		}

		// Reload the image with preshrink
		if err := newImg.Load(c.ImgData, preshrink, 0, 1); err != nil {
			return err
		}
	}

	// Get the geometry of the preshrunk image
	newWidth, newHeight, newAngle, newFlip := ExtractGeometry(
		newImg, c.PO.Rotate(), c.PO.AutoRotate(),
	)

	// Calculate the actual preshrink values
	wpreshrink := float64(c.SrcWidth) / float64(newWidth)
	hpreshrink := float64(c.SrcHeight) / float64(newHeight)

	// If we loaded a thumbnail, check if it's worth using it
	if loadThumbnail {
		// If the thumbnail is not smaller than the original image or
		// if it is shrunk too much, we better keep the original image
		if min(wpreshrink, hpreshrink) <= 1.0 || max(wpreshrink, hpreshrink) > preshrink {
			return nil
		}
	}

	// Swap the image with the preshrunk one and update its orientation in the context
	c.Img.Swap(newImg)
	c.Angle = newAngle
	c.Flip = newFlip

	// Update scales after scale-on-load
	c.WScale *= wpreshrink
	c.HScale *= hpreshrink

	// If preshrink is exact, it's better to set scale to 1.0
	// to prevent additional scaling passes
	if newWidth == imath.Scale(newWidth, c.WScale) {
		c.WScale = 1.0
	}
	if newHeight == imath.Scale(newHeight, c.HScale) {
		c.HScale = 1.0
	}

	// We should crop before scaling, but we scaled the image on load,
	// so we need to adjust crop options
	if c.CropWidth > 0 {
		c.CropWidth = max(1, imath.Shrink(c.CropWidth, wpreshrink))
	}
	if c.CropHeight > 0 {
		c.CropHeight = max(1, imath.Shrink(c.CropHeight, hpreshrink))
	}
	// Adjust crop gravity offsets.
	// We don't need to adjust focus point offsets since they are always relative.
	// For other gravity types, we need to adjust only absolute offsets (>= 1.0 or <= -1.0).
	// We round absolute offsets to prevent turning them to relative (ex: 1.0 => 0.5).
	if c.CropGravity.Type != GravityFocusPoint {
		if math.Abs(c.CropGravity.X) >= 1.0 {
			c.CropGravity.X = math.RoundToEven(c.CropGravity.X / wpreshrink)
		}
		if math.Abs(c.CropGravity.Y) >= 1.0 {
			c.CropGravity.Y = math.RoundToEven(c.CropGravity.Y / hpreshrink)
		}
	}

	return nil
}
