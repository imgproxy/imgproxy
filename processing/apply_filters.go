package processing

func (p *Processor) applyFilters(c *Context) error {
	blur := c.PO.Blur()
	sharpen := c.PO.Sharpen()
	pixelate := c.PO.Pixelate()

	if blur == 0 && sharpen == 0 && pixelate <= 1 {
		return nil
	}

	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	if err := c.Img.RgbColourspace(); err != nil {
		return err
	}

	if err := c.Img.ApplyFilters(blur, sharpen, pixelate); err != nil {
		return err
	}

	return c.Img.CopyMemory()
}
