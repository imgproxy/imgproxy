package processing

func (p *Processor) colorspaceToResult(c *Context) error {
	keepProfile := !c.PO.StripColorProfile() && c.PO.Format().SupportsColourProfile()

	if c.Img.IsLinear() {
		if err := c.Img.RgbColourspace(); err != nil {
			return err
		}
	}

	// vips 8.15+ tends to lose the colour profile during some color conversions.
	// We probably have a backup of the colour profile, so we need to restore it.
	c.Img.RestoreColourProfile()

	if c.Img.ColourProfileImported() {
		if keepProfile {
			// We imported ICC profile and want to keep it,
			// so we need to export it
			if err := c.Img.ExportColourProfile(); err != nil {
				return err
			}
		} else {
			// We imported ICC profile but don't want to keep it,
			// so we need to export image to sRGB for maximum compatibility
			if err := c.Img.ExportColourProfileToSRGB(); err != nil {
				return err
			}
		}
	} else if !keepProfile {
		// We don't import ICC profile and don't want to keep it,
		// so we need to transform it to sRGB for maximum compatibility
		if err := c.Img.TransformColourProfileToSRGB(); err != nil {
			return err
		}
	}

	if !keepProfile {
		return c.Img.RemoveColourProfile()
	}

	return nil
}
