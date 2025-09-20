package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

func rotateAndFlip(ctx *Context) error {
	rotateAngle := options.GetInt(ctx.PO, keys.Rotate, 0)

	if ctx.Angle%360 == 0 && rotateAngle%360 == 0 && !ctx.Flip {
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

	return ctx.Img.Rotate(rotateAngle)
}
