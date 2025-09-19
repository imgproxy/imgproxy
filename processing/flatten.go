package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func flatten(c *Context) error {
	flatten := options.Get(c.PO, keys.Flatten, false)
	format := options.Get(c.PO, keys.Format, imagetype.Unknown)

	if !flatten && format.SupportsAlpha() {
		return nil
	}

	background := options.Get(c.PO, keys.Background, vips.Color{R: 255, G: 255, B: 255})
	return c.Img.Flatten(background)
}
