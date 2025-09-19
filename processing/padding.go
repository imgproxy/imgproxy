package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

func padding(c *Context) error {
	if !options.Get(c.PO, keys.PaddingEnabled, false) {
		return nil
	}

	paddingTop := options.GetInt(c.PO, keys.PaddingTop, 0)
	paddingRight := options.GetInt(c.PO, keys.PaddingRight, 0)
	paddingBottom := options.GetInt(c.PO, keys.PaddingBottom, 0)
	paddingLeft := options.GetInt(c.PO, keys.PaddingLeft, 0)

	paddingTop = imath.ScaleToEven(paddingTop, c.DprScale)
	paddingRight = imath.ScaleToEven(paddingRight, c.DprScale)
	paddingBottom = imath.ScaleToEven(paddingBottom, c.DprScale)
	paddingLeft = imath.ScaleToEven(paddingLeft, c.DprScale)

	return c.Img.Embed(
		c.Img.Width()+paddingLeft+paddingRight,
		c.Img.Height()+paddingTop+paddingBottom,
		paddingLeft,
		paddingTop,
	)
}
