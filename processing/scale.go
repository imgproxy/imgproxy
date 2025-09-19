package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

func scale(c *Context) error {
	if c.WScale == 1 && c.HScale == 1 {
		return nil
	}

	wscale, hscale := c.WScale, c.HScale

	rotateAngle := options.GetInt(c.PO, keys.Rotate, 0)
	if (c.Angle+rotateAngle)%180 == 90 {
		wscale, hscale = hscale, wscale
	}

	return c.Img.Resize(wscale, hscale)
}
