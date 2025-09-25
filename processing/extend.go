package processing

func extendImage(c *Context, width, height int, gravity *GravityOptions) error {
	imgWidth := c.Img.Width()
	imgHeight := c.Img.Height()

	if width <= imgWidth && height <= imgHeight {
		return nil
	}

	if width <= 0 {
		width = imgWidth
	}
	if height <= 0 {
		height = imgHeight
	}

	offX, offY := calcPosition(width, height, imgWidth, imgHeight, gravity, c.DprScale, false)
	return c.Img.Embed(width, height, offX, offY)
}

func (p *Processor) extend(c *Context) error {
	if !c.PO.ExtendEnabled() {
		return nil
	}

	width, height := c.TargetWidth, c.TargetHeight
	gravity := c.PO.ExtendGravity()
	return extendImage(c, width, height, &gravity)
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
	return extendImage(c, width, height, &gravity)
}
