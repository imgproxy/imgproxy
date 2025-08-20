package processing

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func colorspaceToProcessing(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if img.ColourProfileImported() {
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
	} else {
		// vips 8.15+ tends to lose the colour profile during some color conversions.
		// We need to backup the colour profile before the conversion and restore it later.
		img.BackupColourProfile()

		if convertToLinear || !img.IsRGB() {
			if err := img.ImportColourProfile(); err != nil {
				return err
			}
		}
	}

	if convertToLinear {
		return img.LinearColourspace()
	}

	return img.RgbColourspace()
}
