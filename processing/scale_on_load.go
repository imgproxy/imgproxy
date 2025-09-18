package processing

import (
	"math"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func canScaleOnLoad(c *Context, imgdata imagedata.ImageData, scale float64) bool {
	if imgdata == nil || scale == 1 {
		return false
	}

	if imgdata.Format().IsVector() {
		return true
	}

	if c.Config.DisableShrinkOnLoad || scale >= 1 {
		return false
	}

	return imgdata.Format() == imagetype.JPEG ||
		imgdata.Format() == imagetype.WEBP ||
		imgdata.Format() == imagetype.HEIC ||
		imgdata.Format() == imagetype.AVIF
}

func calcJpegShink(shrink float64) int {
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

func scaleOnLoad(c *Context) error {
	wshrink := float64(c.SrcWidth) / float64(imath.Scale(c.SrcWidth, c.WScale))
	hshrink := float64(c.SrcHeight) / float64(imath.Scale(c.SrcHeight, c.HScale))
	preshrink := math.Min(wshrink, hshrink)
	prescale := 1.0 / preshrink

	if c.ImgData != nil && c.ImgData.Format().IsVector() {
		// For vector images, apply the vector base scale
		prescale *= c.VectorBaseScale
	}

	if !canScaleOnLoad(c, c.ImgData, prescale) {
		return nil
	}

	var newWidth, newHeight int

	if c.ImgData.Format().SupportsThumbnail() {
		thumbnail := new(vips.Image)
		defer thumbnail.Clear()

		if err := thumbnail.LoadThumbnail(c.ImgData); err != nil {
			log.Debugf("Can't load thumbnail: %s", err)
			return nil
		}

		angle, flip := 0, false
		newWidth, newHeight, angle, flip = extractMeta(thumbnail, c.PO.Rotate, c.PO.AutoRotate)

		if newWidth >= c.SrcWidth || float64(newWidth)/float64(c.SrcWidth) < prescale {
			return nil
		}

		c.Img.Swap(thumbnail)
		c.Angle = angle
		c.Flip = flip
	} else {
		jpegShrink := calcJpegShink(preshrink)

		if c.ImgData.Format() == imagetype.JPEG && jpegShrink == 1 {
			return nil
		}

		if err := c.Img.Load(c.ImgData, jpegShrink, prescale, 1); err != nil {
			return err
		}

		newWidth, newHeight, _, _ = extractMeta(c.Img, c.PO.Rotate, c.PO.AutoRotate)
	}

	// Update scales after scale-on-load
	wpreshrink := float64(c.SrcWidth) / float64(newWidth)
	hpreshrink := float64(c.SrcHeight) / float64(newHeight)

	c.WScale = wpreshrink * c.WScale
	if newWidth == imath.Scale(newWidth, c.WScale) {
		c.WScale = 1.0
	}

	c.HScale = hpreshrink * c.HScale
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
	if c.CropGravity.Type != options.GravityFocusPoint {
		// Adjust only when crop gravity offsets are absolute
		if math.Abs(c.CropGravity.X) >= 1.0 {
			// Round offsets to prevent turning absolute offsets to relative (ex: 1.0 => 0.5)
			c.CropGravity.X = math.RoundToEven(c.CropGravity.X / wpreshrink)
		}
		if math.Abs(c.CropGravity.Y) >= 1.0 {
			c.CropGravity.Y = math.RoundToEven(c.CropGravity.Y / hpreshrink)
		}
	}

	return nil
}
