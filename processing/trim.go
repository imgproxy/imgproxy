package processing

import (
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/vips"
)

func trim(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Trim.Enabled {
		return nil
	}

	if err := img.Trim(po.Trim.Threshold, po.Trim.Smart, po.Trim.Color, po.Trim.EqualHor, po.Trim.EqualVer); err != nil {
		return err
	}
	if err := copyMemoryAndCheckTimeout(pctx.ctx, img); err != nil {
		return err
	}

	pctx.trimmed = true

	return nil
}
