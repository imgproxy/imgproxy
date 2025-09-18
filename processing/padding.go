package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
)

func (p *Processor) padding(c *Context) error {
	if !c.PO.Padding.Enabled {
		return nil
	}

	paddingTop := imath.ScaleToEven(c.PO.Padding.Top, c.DprScale)
	paddingRight := imath.ScaleToEven(c.PO.Padding.Right, c.DprScale)
	paddingBottom := imath.ScaleToEven(c.PO.Padding.Bottom, c.DprScale)
	paddingLeft := imath.ScaleToEven(c.PO.Padding.Left, c.DprScale)

	return c.Img.Embed(
		c.Img.Width()+paddingLeft+paddingRight,
		c.Img.Height()+paddingTop+paddingBottom,
		paddingLeft,
		paddingTop,
	)
}
