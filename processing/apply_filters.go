package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func applyFilters(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.Blur == 0 && po.Sharpen == 0 && po.Pixelate <= 1 {
		return nil
	}

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	if err := img.ApplyFilters(po.Blur, po.Sharpen, po.Pixelate); err != nil {
		return err
	}

	return img.CopyMemory()
}
