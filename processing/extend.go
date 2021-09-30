package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extend(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	resultWidth := imath.Scale(po.Width, po.Dpr)
	resultHeight := imath.Scale(po.Height, po.Dpr)

	if !po.Extend.Enabled || (resultWidth <= img.Width() && resultHeight <= img.Height()) {
		return nil
	}

	offX, offY := calcPosition(resultWidth, resultHeight, img.Width(), img.Height(), &po.Extend.Gravity, false)
	return img.Embed(resultWidth, resultHeight, offX, offY)
}
