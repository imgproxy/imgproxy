package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extendImage(img *vips.Image, width, height int, gravity *options.GravityOptions, offsetScale float64) error {
	imgWidth := img.Width()
	imgHeight := img.Height()

	if width <= imgWidth && height <= imgHeight {
		return nil
	}

	if width <= 0 {
		width = imgWidth
	}
	if height <= 0 {
		height = imgHeight
	}

	offX, offY := calcPosition(width, height, imgWidth, imgHeight, gravity, offsetScale, false)
	return img.Embed(width, height, offX, offY)
}

func extend(c *Context) error {
	if !options.Get(c.PO, keys.ExtendEnabled, false) {
		return nil
	}

	width, height := c.TargetWidth, c.TargetHeight
	gravity := options.GetGravity(c.PO, keys.ExtendGravity, options.GravityCenter)
	return extendImage(c.Img, width, height, &gravity, c.DprScale)
}

func extendAspectRatio(c *Context) error {
	if !options.Get(c.PO, keys.ExtendAspectRatioEnabled, false) {
		return nil
	}

	width, height := c.ExtendAspectRatioWidth, c.ExtendAspectRatioHeight
	if width == 0 || height == 0 {
		return nil
	}

	gravity := options.GetGravity(c.PO, keys.ExtendAspectRatioGravity, options.GravityCenter)
	return extendImage(c.Img, width, height, &gravity, c.DprScale)
}
