package processing

func trim(c *Context) error {
	if !c.PO.Trim.Enabled {
		return nil
	}

	// We need to import color profile before trim
	if err := colorspaceToProcessing(c); err != nil {
		return err
	}

	if err := c.Img.Trim(c.PO.Trim.Threshold, c.PO.Trim.Smart, c.PO.Trim.Color, c.PO.Trim.EqualHor, c.PO.Trim.EqualVer); err != nil {
		return err
	}
	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	c.ImgData = nil

	return nil
}
