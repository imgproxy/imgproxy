package processing

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func importColorProfile(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if pctx.iccImported {
		return nil
	}

	if err := img.Rad2Float(); err != nil {
		return err
	}

	convertToLinear := config.UseLinearColorspace && (pctx.wscale != 1 || pctx.hscale != 1)

	if img.IsLinear() {
		// The image is linear. If we keep its ICC, we'll get wrong colors after
		// converting it to sRGB
		img.RemoveColourProfile()
	} else if convertToLinear || !img.IsRGB() {
		if err := img.ImportColourProfile(); err != nil {
			return err
		}
		pctx.iccImported = true
	}

	if convertToLinear {
		return img.LinearColourspace()
	}

	return img.RgbColourspace()
}
