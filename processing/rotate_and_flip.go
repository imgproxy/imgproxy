package processing

func (p *Processor) rotateAndFlip(c *Context) error {
	rotateAngle := c.PO.Rotate()
	flipX := c.PO.FlipHorizontal()
	flipY := c.PO.FlipVertical()

	shouldRotate := c.Angle%360 != 0 || rotateAngle%360 != 0
	shouldFlip := c.Flip || flipX || flipY

	if !shouldRotate && !shouldFlip {
		return nil
	}

	// We need the image in random access mode, so we copy it to memory.
	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	// Rotate according to EXIF orientation
	if err := c.Img.Rotate(c.Angle); err != nil {
		return err
	}

	// Flip according to EXIF orientation
	if err := c.Img.Flip(c.Flip, false); err != nil {
		return err
	}

	// Rotate according to user-specified options
	if err := c.Img.Rotate(rotateAngle); err != nil {
		return err
	}

	// Flip according to user-specified options
	return c.Img.Flip(flipX, flipY)
}
