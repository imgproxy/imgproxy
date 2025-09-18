package processing

func rotateAndFlip(ctx *Context) error {
	if ctx.Angle%360 == 0 && ctx.PO.Rotate%360 == 0 && !ctx.Flip {
		return nil
	}

	if err := ctx.Img.CopyMemory(); err != nil {
		return err
	}

	if err := ctx.Img.Rotate(ctx.Angle); err != nil {
		return err
	}

	if ctx.Flip {
		if err := ctx.Img.Flip(); err != nil {
			return err
		}
	}

	return ctx.Img.Rotate(ctx.PO.Rotate)
}
