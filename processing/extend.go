package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
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

func (p *Processor) extend(c *Context) error {
	if !c.PO.Extend.Enabled {
		return nil
	}

	width, height := c.TargetWidth, c.TargetHeight
	return extendImage(c.Img, width, height, &c.PO.Extend.Gravity, c.DprScale)
}

func (p *Processor) extendAspectRatio(c *Context) error {
	if !c.PO.ExtendAspectRatio.Enabled {
		return nil
	}

	width, height := c.ExtendAspectRatioWidth, c.ExtendAspectRatioHeight
	if width == 0 || height == 0 {
		return nil
	}

	return extendImage(c.Img, width, height, &c.PO.ExtendAspectRatio.Gravity, c.DprScale)
}
