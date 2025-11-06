package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func rotateAndFlip(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	shouldRotate := pctx.angle%360 != 0 || po.Rotate%360 != 0
	shouldFlip := pctx.flip || po.Flip.Horizontal || po.Flip.Vertical

	if !shouldRotate && !shouldFlip {
		return nil
	}

	// We need the image in random access mode, so we copy it to memory.
	if err := img.CopyMemory(); err != nil {
		return err
	}

	// Rotate according to EXIF orientation
	if err := img.Rotate(pctx.angle); err != nil {
		return err
	}

	// Flip according to EXIF orientation
	if err := img.Flip(pctx.flip, false); err != nil {
		return err
	}

	// Rotate according to user-specified options
	if err := img.Rotate(po.Rotate); err != nil {
		return err
	}

	// Flip according to user-specified options
	return img.Flip(po.Flip.Horizontal, po.Flip.Vertical)
}
