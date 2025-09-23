package processing

import (
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extendImage(img *vips.Image, width, height int, gravity *GravityOptions, offsetScale float64) error {
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

func (p *Processor) extend(c *Context) error {
	if !c.PO.ExtendEnabled() {
		return nil
	}

	width, height := c.TargetWidth, c.TargetHeight
	gravity := c.PO.ExtendGravity()
	return extendImage(c.Img, width, height, &gravity, c.DprScale)
}

func (p *Processor) extendAspectRatio(c *Context) error {
	if !c.PO.ExtendAspectRatioEnabled() {
		return nil
	}

	width, height := c.ExtendAspectRatioWidth, c.ExtendAspectRatioHeight
	if width == 0 || height == 0 {
		return nil
	}

	gravity := c.PO.ExtendAspectRatioGravity()
	return extendImage(c.Img, width, height, &gravity, c.DprScale)
}
