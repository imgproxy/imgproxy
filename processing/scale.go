package processing

import (
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/vips"
)

func scale(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if pctx.wscale != 1 || pctx.hscale != 1 {
		if err := img.Resize(pctx.wscale, pctx.hscale); err != nil {
			return err
		}
	}

	return copyMemoryAndCheckTimeout(pctx.ctx, img)
}
