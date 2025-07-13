package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extendImage(img *vips.Image, width, height int, gravity *options.GravityOptions, offsetScale float64) error {
	imgWidth := img.Width()
	imgHeight := img.Height()

	if width <= imgWidth && height <= imgHeight {
		return nil
	}

	if width <= 0 {
		width = imgWidth
	}
	if height <= 0 {
		height = imgHeight
	}

	offX, offY := calcPosition(width, height, imgWidth, imgHeight, gravity, offsetScale, false)
	return img.Embed(width, height, offX, offY)
}

func extend(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.Extend.Enabled {
		return nil
	}

	width, height := pctx.targetWidth, pctx.targetHeight
	return extendImage(img, width, height, &po.Extend.Gravity, pctx.dprScale)
}

func extendAspectRatio(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if !po.ExtendAspectRatio.Enabled {
		return nil
	}

	width, height := pctx.extendAspectRatioWidth, pctx.extendAspectRatioHeight
	if width == 0 || height == 0 {
		return nil
	}

	return extendImage(img, width, height, &po.ExtendAspectRatio.Gravity, pctx.dprScale)
}

func extendContain(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.ResizingType != options.ResizeContain {
		return nil
	}

	if pctx.targetWidth == 0 || pctx.targetHeight == 0 {
		return nil
	}

	imgWidth := img.Width()
	imgHeight := img.Height()

	if imgWidth >= pctx.targetWidth && imgHeight >= pctx.targetHeight {
		return nil
	}

	return extendImage(img, pctx.targetWidth, pctx.targetHeight, &po.Gravity, pctx.dprScale)
}
