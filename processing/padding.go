package processing

import (
	"github.com/imgproxy/imgproxy/v2/imagedata"
	"github.com/imgproxy/imgproxy/v2/imath"
	"github.com/imgproxy/imgproxy/v2/options"
	"github.com/imgproxy/imgproxy/v2/vips"
)

func padding(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Padding.Enabled {
		return nil
	}

	paddingTop := imath.Scale(po.Padding.Top, po.Dpr)
	paddingRight := imath.Scale(po.Padding.Right, po.Dpr)
	paddingBottom := imath.Scale(po.Padding.Bottom, po.Dpr)
	paddingLeft := imath.Scale(po.Padding.Left, po.Dpr)

	return img.Embed(
		img.Width()+paddingLeft+paddingRight,
		img.Height()+paddingTop+paddingBottom,
		paddingLeft,
		paddingTop,
	)
}
