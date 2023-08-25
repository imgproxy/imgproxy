package processing

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func applyBlurRegions(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {

	if err := img.CopyMemory(); err != nil {
		return err
	}

	if err := img.RgbColourspace(); err != nil {
		return err
	}

	for _, reg := range po.BlurRegions {
		fmt.Printf("BLUR REGION: %+v\n", reg)

		if err := img.BlurRegion(reg.X0, reg.Y0, reg.X1, reg.Y1, reg.Sigma); err != nil {
			return err
		}
	}

	return img.CopyMemory()
}
