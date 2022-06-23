package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
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

	// When image has alpha, we need to premultiply it to get rid of black edges
	if err := img.Premultiply(); err != nil {
		return err
	}

	if po.Blur > 0 {
		if err := img.Blur(po.Blur); err != nil {
			return err
		}
	}

	if po.Sharpen > 0 {
		if err := img.Sharpen(po.Sharpen); err != nil {
			return err
		}
	}

	if po.Pixelate > 1 {
		pixels := imath.Min(po.Pixelate, imath.Min(img.Width(), img.Height()))
		if err := img.Pixelate(pixels); err != nil {
			return err
		}
	}

	if err := img.Unpremultiply(); err != nil {
		return err
	}

	if err := img.CastUchar(); err != nil {
		return err
	}

	return img.CopyMemory()
}
