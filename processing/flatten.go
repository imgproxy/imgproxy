package processing

func (p *Processor) flatten(c *Context) error {
	if !c.PO.Flatten && c.PO.Format.SupportsAlpha() {
		return nil
	}

	return c.Img.Flatten(c.PO.Background)
}
