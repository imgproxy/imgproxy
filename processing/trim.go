package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func trim(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Trim.Enabled {
		return nil
	}

	// We need to import color profile before trim
	if err := importColorProfile(pctx, img, po, imgdata); err != nil {
		return err
	}

	if err := img.Trim(po.Trim.Threshold, po.Trim.Smart, po.Trim.Color, po.Trim.EqualHor, po.Trim.EqualVer); err != nil {
		return err
	}
	if err := img.CopyMemory(); err != nil {
		return err
	}

	pctx.trimmed = true

	return nil
}
