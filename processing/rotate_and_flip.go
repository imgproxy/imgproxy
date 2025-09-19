package processing

func (p *Processor) rotateAndFlip(c *Context) error {
	if c.Angle%360 == 0 && c.PO.Rotate%360 == 0 && !c.Flip {
		return nil
	}

	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	if err := c.Img.Rotate(c.Angle); err != nil {
		return err
	}

	if c.Flip {
		if err := c.Img.Flip(); err != nil {
			return err
		}
	}

	return c.Img.Rotate(c.PO.Rotate)
}
