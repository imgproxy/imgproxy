package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func padding(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Padding.Enabled {
		return nil
	}

	paddingTop := imath.ScaleToEven(po.Padding.Top, pctx.dprScale)
	paddingRight := imath.ScaleToEven(po.Padding.Right, pctx.dprScale)
	paddingBottom := imath.ScaleToEven(po.Padding.Bottom, pctx.dprScale)
	paddingLeft := imath.ScaleToEven(po.Padding.Left, pctx.dprScale)

	return img.Embed(
		img.Width()+paddingLeft+paddingRight,
		img.Height()+paddingTop+paddingBottom,
		paddingLeft,
		paddingTop,
	)
}
