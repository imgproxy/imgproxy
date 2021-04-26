package processing

import (
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/vips"
)

func rotateAndFlip(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
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
