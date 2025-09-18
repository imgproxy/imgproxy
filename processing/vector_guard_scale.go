package processing

import (
	"math"
)

// vectorGuardScale checks if the image is a vector format and downscales it
// to the maximum allowed resolution if necessary
func (p *Processor) vectorGuardScale(c *Context) error {
	if c.ImgData == nil || !c.ImgData.Format().IsVector() {
		return nil
	}

	if resolution := c.Img.Width() * c.Img.Height(); resolution > c.PO.SecurityOptions.MaxSrcResolution {
		scale := math.Sqrt(float64(c.PO.SecurityOptions.MaxSrcResolution) / float64(resolution))
		c.VectorBaseScale = scale

		if err := c.Img.Load(c.ImgData, 1, scale, 1); err != nil {
			return err
		}
	}
	c.CalcParams()

	return nil
}
