package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
)

func (p *Processor) padding(c *Context) error {
	if !c.PO.PaddingEnabled() {
		return nil
	}

	paddingTop := imath.ScaleToEven(c.PO.PaddingTop(), c.DprScale)
	paddingRight := imath.ScaleToEven(c.PO.PaddingRight(), c.DprScale)
	paddingBottom := imath.ScaleToEven(c.PO.PaddingBottom(), c.DprScale)
	paddingLeft := imath.ScaleToEven(c.PO.PaddingLeft(), c.DprScale)

	return c.Img.Embed(
		c.Img.Width()+paddingLeft+paddingRight,
		c.Img.Height()+paddingTop+paddingBottom,
		paddingLeft,
		paddingTop,
	)
}
