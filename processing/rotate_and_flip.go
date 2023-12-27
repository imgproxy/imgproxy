package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func rotateAndFlip(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if pctx.angle%360 == 0 && po.Rotate%360 == 0 && !pctx.flip {
		return nil
	}

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if err := img.Rotate(pctx.angle); err != nil {
		return err
	}

	if pctx.flip {
		if err := img.Flip(); err != nil {
			return err
		}
	}

	return img.Rotate(po.Rotate)
}
