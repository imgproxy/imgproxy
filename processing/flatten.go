package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func flatten(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedatanew.ImageData) error {
	if !po.Flatten && po.Format.SupportsAlpha() {
		return nil
	}

	return img.Flatten(po.Background)
}
