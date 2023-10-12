package processing

import (
	"math"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func canScaleOnLoad(pctx *pipelineContext, imgdata *imagedata.ImageData, scale float64) bool {
	if imgdata == nil || pctx.trimmed || scale == 1 {
		return false
	}

	if imgdata.Type == imagetype.SVG {
		return true
	}

	if config.DisableShrinkOnLoad || scale >= 1 {
		return false
	}

	return imgdata.Type == imagetype.JPEG ||
		imgdata.Type == imagetype.WEBP ||
		imgdata.Type == imagetype.HEIC ||
		imgdata.Type == imagetype.AVIF
}

func calcJpegShink(scale float64, imgtype imagetype.Type) int {
	shrink := int(1.0 / scale)

	switch {
	case shrink >= 8:
		return 8
	case shrink >= 4:
		return 4
	case shrink >= 2:
		return 2
	}

	return 1
}

func scaleOnLoad(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	prescale := math.Max(pctx.wscale, pctx.hscale)

	if !canScaleOnLoad(pctx, imgdata, prescale) {
		return nil
	}

	var newWidth, newHeight int

	if imgdata.Type.SupportsThumbnail() {
		thumbnail := new(vips.Image)
		defer thumbnail.Clear()

		if err := thumbnail.LoadThumbnail(imgdata); err != nil {
			log.Debugf("Can't load thumbnail: %s", err)
			return nil
		}

		angle, flip := 0, false
		newWidth, newHeight, angle, flip = extractMeta(thumbnail, po.Rotate, po.AutoRotate)

		if newWidth >= pctx.srcWidth || float64(newWidth)/float64(pctx.srcWidth) < prescale {
			return nil
		}

		img.Swap(thumbnail)
		pctx.angle = angle
		pctx.flip = flip
	} else {
		jpegShrink := calcJpegShink(prescale, pctx.imgtype)

		if pctx.imgtype == imagetype.JPEG && jpegShrink == 1 {
			return nil
		}

		if err := img.Load(imgdata, jpegShrink, prescale, 1); err != nil {
			return err
		}

		newWidth, newHeight, _, _ = extractMeta(img, po.Rotate, po.AutoRotate)
	}

	// Update scales after scale-on-load
	wpreshrink := float64(pctx.srcWidth) / float64(newWidth)
	hpreshrink := float64(pctx.srcHeight) / float64(newHeight)

	pctx.wscale = wpreshrink * pctx.wscale
	if newWidth == imath.Scale(newWidth, pctx.wscale) {
		pctx.wscale = 1.0
	}

	pctx.hscale = hpreshrink * pctx.hscale
	if newHeight == imath.Scale(newHeight, pctx.hscale) {
		pctx.hscale = 1.0
	}

	if pctx.cropWidth > 0 {
		pctx.cropWidth = imath.Max(1, imath.Shrink(pctx.cropWidth, wpreshrink))
	}
	if pctx.cropHeight > 0 {
		pctx.cropHeight = imath.Max(1, imath.Shrink(pctx.cropHeight, hpreshrink))
	}
	if pctx.cropGravity.Type != options.GravityFocusPoint {
		pctx.cropGravity.X /= wpreshrink
		pctx.cropGravity.Y /= hpreshrink
	}

	return nil
}
