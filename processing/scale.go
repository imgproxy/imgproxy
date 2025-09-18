package processing

func scale(c *Context) error {
	if c.WScale == 1 && c.HScale == 1 {
		return nil
	}

	wscale, hscale := c.WScale, c.HScale
	if (c.Angle+c.PO.Rotate)%180 == 90 {
		wscale, hscale = hscale, wscale
	}

	return c.Img.Resize(wscale, hscale)
}
