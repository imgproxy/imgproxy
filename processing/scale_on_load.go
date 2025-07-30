package processing

import (
	"math"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func canScaleOnLoad(pctx *pipelineContext, imgdata imagedatanew.ImageData, scale float64) bool {
	if imgdata == nil || pctx.trimmed || scale == 1 {
		return false
	}

	if imgdata.Format().IsVector() {
		return true
	}

	if config.DisableShrinkOnLoad || scale >= 1 {
		return false
	}

	return imgdata.Format() == imagetype.JPEG ||
		imgdata.Format() == imagetype.WEBP ||
		imgdata.Format() == imagetype.HEIC ||
		imgdata.Format() == imagetype.AVIF
}

func calcJpegShink(shrink float64) int {
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

func scaleOnLoad(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedatanew.ImageData) error {
	wshrink := float64(pctx.srcWidth) / float64(imath.Scale(pctx.srcWidth, pctx.wscale))
	hshrink := float64(pctx.srcHeight) / float64(imath.Scale(pctx.srcHeight, pctx.hscale))
	preshrink := math.Min(wshrink, hshrink)
	prescale := 1.0 / preshrink

	if !canScaleOnLoad(pctx, imgdata, prescale) {
		return nil
	}

	var newWidth, newHeight int

	if imgdata.Format().SupportsThumbnail() {
		thumbnail := new(vips.Image)
		defer thumbnail.Clear()

		if err := thumbnail.LoadThumbnail(imagedata.From(imgdata)); err != nil {
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
		jpegShrink := calcJpegShink(preshrink)

		if pctx.imgtype == imagetype.JPEG && jpegShrink == 1 {
			return nil
		}

		if err := img.Load(imagedata.From(imgdata), jpegShrink, prescale, 1); err != nil {
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

	// We should crop before scaling, but we scaled the image on load,
	// so we need to adjust crop options
	if pctx.cropWidth > 0 {
		pctx.cropWidth = imath.Max(1, imath.Shrink(pctx.cropWidth, wpreshrink))
	}
	if pctx.cropHeight > 0 {
		pctx.cropHeight = imath.Max(1, imath.Shrink(pctx.cropHeight, hpreshrink))
	}
	if pctx.cropGravity.Type != options.GravityFocusPoint {
		// Adjust only when crop gravity offsets are absolute
		if math.Abs(pctx.cropGravity.X) >= 1.0 {
			// Round offsets to prevent turning absolute offsets to relative (ex: 1.0 => 0.5)
			pctx.cropGravity.X = math.RoundToEven(pctx.cropGravity.X / wpreshrink)
		}
		if math.Abs(pctx.cropGravity.Y) >= 1.0 {
			pctx.cropGravity.Y = math.RoundToEven(pctx.cropGravity.Y / hpreshrink)
		}
	}

	return nil
}
