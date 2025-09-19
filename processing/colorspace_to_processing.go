package processing

func (p *Processor) colorspaceToProcessing(c *Context) error {
	if c.Img.ColourProfileImported() {
		return nil
	}

	if err := c.Img.Rad2Float(); err != nil {
		return err
	}

	convertToLinear := p.config.UseLinearColorspace && (c.WScale != 1 || c.HScale != 1)

	if c.Img.IsLinear() {
		// The image is linear. If we keep its ICC, we'll get wrong colors after
		// converting it to sRGB
		c.Img.RemoveColourProfile()
	} else {
		// vips 8.15+ tends to lose the colour profile during some color conversions.
		// We need to backup the colour profile before the conversion and restore it later.
		c.Img.BackupColourProfile()

		if convertToLinear || !c.Img.IsRGB() {
			if err := c.Img.ImportColourProfile(); err != nil {
				return err
			}
		}
	}

	if convertToLinear {
		return c.Img.LinearColourspace()
	}

	return c.Img.RgbColourspace()
}
