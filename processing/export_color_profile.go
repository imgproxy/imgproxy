package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func exportColorProfile(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	keepProfile := !po.StripColorProfile && po.Format.SupportsColourProfile()

	if img.IsLinear() {
		if err := img.RgbColourspace(); err != nil {
			return err
		}
	}

	if pctx.iccImported {
		if keepProfile {
			// We imported ICC profile and want to keep it,
			// so we need to export it
			if err := img.ExportColourProfile(); err != nil {
				return err
			}
		} else {
			// We imported ICC profile but don't want to keep it,
			// so we need to export image to sRGB for maximum compatibility
			if err := img.ExportColourProfileToSRGB(); err != nil {
				return err
			}
		}
	} else if !keepProfile {
		// We don't import ICC profile and don't want to keep it,
		// so we need to transform it to sRGB for maximum compatibility
		if err := img.TransformColourProfile(); err != nil {
			return err
		}
	}

	if !keepProfile {
		return img.RemoveColourProfile()
	}

	return nil
}
