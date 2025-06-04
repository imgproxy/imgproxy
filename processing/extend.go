package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extendImage(img *vips.Image, resultWidth, resultHeight int, opts *options.ExtendOptions, offsetScale float64, extendAr bool) error {
	imgWidth := img.Width()
	imgHeight := img.Height()

	if !opts.Enabled || (resultWidth <= imgWidth && resultHeight <= imgHeight) {
		return nil
	}

	if resultWidth <= 0 {
		if extendAr {
			return nil
		}
		resultWidth = imgWidth
	}
	if resultHeight <= 0 {
		if extendAr {
			return nil
		}
		resultHeight = imgHeight
	}

	if extendAr && resultWidth > imgWidth && resultHeight > imgHeight {
		diffW := float64(resultWidth) / float64(imgWidth)
		diffH := float64(resultHeight) / float64(imgHeight)

		switch {
		case diffH > diffW:
			resultHeight = imath.Scale(imgWidth, float64(resultHeight)/float64(resultWidth))
			resultWidth = imgWidth

		case diffW > diffH:
			resultWidth = imath.Scale(imgHeight, float64(resultWidth)/float64(resultHeight))
			resultHeight = imgHeight

		default:
			// The image has the requested arpect ratio
			return nil
		}
	}

	offX, offY := calcPosition(resultWidth, resultHeight, imgWidth, imgHeight, &opts.Gravity, offsetScale, false)
	return img.Embed(resultWidth, resultHeight, offX, offY)
}

func extend(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	resultWidth, resultHeight := resultSize(po, pctx.dprScale)
	return extendImage(img, resultWidth, resultHeight, &po.Extend, pctx.dprScale, false)
}

func extendAspectRatio(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	resultWidth, resultHeight := resultSize(po, pctx.dprScale)
	return extendImage(img, resultWidth, resultHeight, &po.ExtendAspectRatio, pctx.dprScale, true)
}
