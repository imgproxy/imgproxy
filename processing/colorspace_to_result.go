package processing

func (p *Processor) colorspaceToResult(c *Context) error {
	keepProfile := !c.PO.StripColorProfile() && c.PO.Format().SupportsColourProfile()
	profileImported := c.Img.ColourProfileImported()

	// vips 8.15+ tends to lose the colour profile during some color conversions.
	// We probably have a backup of the colour profile, so we need to restore it.
	c.Img.RestoreColourProfile()

	// NOTE:
	// If we imported ICC but don't want to keep it, we can just remove it without transforming,
	// because the image is already in a standard color space.
	// If we didn't import ICC but want to keep it, we can just keep it without exporting,
	// because the image is still in the ICC's color space.

	switch {
	case keepProfile && profileImported:
		// We imported ICC profile and want to keep it,
		// so we need to export it
		if err := c.Img.ExportColourProfile(); err != nil {
			return err
		}
	case !keepProfile && !profileImported:
		// We didn't import ICC profile and don't want to keep it,
		// so we need to transform it to sRGB/sGray for maximum compatibility
		if err := c.Img.TransformColourProfileToStandard(); err != nil {
			return err
		}
	}

	if !keepProfile {
		return c.Img.RemoveColourProfile()
	}

	return nil
}
