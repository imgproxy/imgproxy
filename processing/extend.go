package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extendImage(img *vips.Image, resultWidth, resultHeight int, opts *options.ExtendOptions, extendAr bool) error {
	if !opts.Enabled || (resultWidth <= img.Width() && resultHeight <= img.Height()) {
		return nil
	}

	if extendAr && resultWidth > img.Width() && resultHeight > img.Height() {
		diffW := float64(resultWidth) / float64(img.Width())
		diffH := float64(resultHeight) / float64(img.Height())

		switch {
		case diffH > diffW:
			resultHeight = imath.Scale(img.Width(), float64(resultHeight)/float64(resultWidth))
			resultWidth = img.Width()

		case diffW > diffH:
			resultWidth = imath.Scale(img.Height(), float64(resultWidth)/float64(resultHeight))
			resultHeight = img.Height()

		default:
			// The image has the requested arpect ratio
			return nil
		}
	}

	offX, offY := calcPosition(resultWidth, resultHeight, img.Width(), img.Height(), &opts.Gravity, false)
	return img.Embed(resultWidth, resultHeight, offX, offY)
}

func extend(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	resultWidth, resultHeight := resultSize(po)
	return extendImage(img, resultWidth, resultHeight, &po.Extend, false)
}

func extendAspectRatio(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	resultWidth, resultHeight := resultSize(po)
	return extendImage(img, resultWidth, resultHeight, &po.ExtendAspectRatio, true)
}
