package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

const (
	// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/master/src/webp/encode.h#529
	webpMaxDimension = 16383.0
	gifMaxDimension  = 65535.0
	icoMaxDimension  = 256.0
)

func fixWebpSize(img *vips.Image) error {
	webpLimitShrink := float64(imath.Max(img.Width(), img.Height())) / webpMaxDimension

	if webpLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / webpLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	log.Warningf("WebP dimension size is limited to %d. The image is rescaled to %dx%d", int(webpMaxDimension), img.Width(), img.Height())

	return nil
}

func fixGifSize(img *vips.Image) error {
	gifMaxResolution := float64(vips.GifResolutionLimit())
	gifResLimitShrink := float64(img.Width()*img.Height()) / gifMaxResolution
	gifDimLimitShrink := float64(imath.Max(img.Width(), img.Height())) / gifMaxDimension

	gifLimitShrink := math.Max(gifResLimitShrink, gifDimLimitShrink)

	if gifLimitShrink <= 1.0 {
		return nil
	}

	scale := math.Sqrt(1.0 / gifLimitShrink)
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	log.Warningf("GIF resolution is limited to %d and dimension size is limited to %d. The image is rescaled to %dx%d", int(gifMaxResolution), int(gifMaxDimension), img.Width(), img.Height())

	return nil
}

func fixIcoSize(img *vips.Image) error {
	icoLimitShrink := float64(imath.Max(img.Width(), img.Height())) / icoMaxDimension

	if icoLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / icoLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	log.Warningf("ICO dimension size is limited to %d. The image is rescaled to %dx%d", int(icoMaxDimension), img.Width(), img.Height())

	return nil
}

func fixSize(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	switch po.Format {
	case imagetype.WEBP:
		return fixWebpSize(img)
	case imagetype.GIF:
		return fixGifSize(img)
	case imagetype.ICO:
		return fixIcoSize(img)
	}

	return nil
}
