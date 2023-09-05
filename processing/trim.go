package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func trim(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Trim.Enabled {
		return nil
	}

	// The size of a vector image are not checked during download, yet it can be very large.
	// So we should scale it down to the maximum allowed resolution
	if imgdata != nil && imgdata.Type.IsVector() {
		if resolution := img.Width() * img.Height(); resolution > po.SecurityOptions.MaxSrcResolution {
			scale := math.Sqrt(float64(po.SecurityOptions.MaxSrcResolution) / float64(resolution))
			if err := img.Load(imgdata, 1, scale, 1); err != nil {
				return err
			}
		}
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
