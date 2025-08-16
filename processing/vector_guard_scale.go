package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// vectorGuardScale checks if the image is a vector format and downscales it
// to the maximum allowed resolution if necessary
func vectorGuardScale(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedata.ImageData) error {
	if imgdata == nil || !imgdata.Format().IsVector() {
		return nil
	}

	if resolution := img.Width() * img.Height(); resolution > po.SecurityOptions.MaxSrcResolution {
		scale := math.Sqrt(float64(po.SecurityOptions.MaxSrcResolution) / float64(resolution))
		pctx.vectorBaseScale = scale

		if err := img.Load(imgdata, 1, scale, 1); err != nil {
			return err
		}
	}

	return nil
}
