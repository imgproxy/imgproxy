package processing

import (
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/master/src/webp/encode.h#529
const webpMaxDimension = 16383.0

func fixWebpSize(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.Format != imagetype.WEBP {
		return nil
	}
	webpLimitShrink := float64(imath.Max(img.Width(), img.Height())) / webpMaxDimension

	if webpLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / webpLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	log.Warningf("WebP dimension size is limited to %d. The image is rescaled to %dx%d", int(webpMaxDimension), img.Width(), img.Height())

	return img.CopyMemory()
}
