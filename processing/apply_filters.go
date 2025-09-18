package processing

func applyFilters(c *Context) error {
	if c.PO.Blur == 0 && c.PO.Sharpen == 0 && c.PO.Pixelate <= 1 {
		return nil
	}

	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	if err := c.Img.RgbColourspace(); err != nil {
		return err
	}

	if err := c.Img.ApplyFilters(c.PO.Blur, c.PO.Sharpen, c.PO.Pixelate); err != nil {
		return err
	}

	return c.Img.CopyMemory()
}
