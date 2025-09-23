package processing

func (p *Processor) trim(c *Context) error {
	if !c.PO.TrimEnabled() {
		return nil
	}

	// We need to import color profile before trim
	if err := p.colorspaceToProcessing(c); err != nil {
		return err
	}

	if err := c.Img.Trim(
		c.PO.TrimThreshold(),
		c.PO.TrimSmart(),
		c.PO.TrimColor(),
		c.PO.TrimEqualHor(),
		c.PO.TrimEqualVer(),
	); err != nil {
		return err
	}
	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	c.ImgData = nil
	c.CalcParams()

	return nil
}
