package processing

import (
	"fmt"
	"log/slog"
	"math"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
)

const (
	// https://chromium.googlesource.com/webm/libwebp/+/refs/heads/master/src/webp/encode.h#529
	webpMaxDimension = 16383.0
	heifMaxDimension = 16384.0
	gifMaxDimension  = 65535.0
	icoMaxDimension  = 256.0
)

func fixWebpSize(img *vips.Image) error {
	webpLimitShrink := float64(max(img.Width(), img.Height())) / webpMaxDimension

	if webpLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / webpLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	slog.Warn(fmt.Sprintf(
		"WebP dimension size is limited to %d. The image is rescaled to %dx%d",
		int(webpMaxDimension), img.Width(), img.Height(),
	))

	return nil
}

func fixHeifSize(img *vips.Image) error {
	heifLimitShrink := float64(max(img.Width(), img.Height())) / heifMaxDimension

	if heifLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / heifLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	slog.Warn(fmt.Sprintf(
		"AVIF/HEIC dimension size is limited to %d. The image is rescaled to %dx%d",
		int(heifMaxDimension), img.Width(), img.Height(),
	))

	return nil
}

func fixGifSize(img *vips.Image) error {
	gifMaxResolution := float64(vips.GifResolutionLimit())
	gifResLimitShrink := float64(img.Width()*img.Height()) / gifMaxResolution
	gifDimLimitShrink := float64(max(img.Width(), img.Height())) / gifMaxDimension

	gifLimitShrink := math.Max(gifResLimitShrink, gifDimLimitShrink)

	if gifLimitShrink <= 1.0 {
		return nil
	}

	scale := math.Sqrt(1.0 / gifLimitShrink)
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	slog.Warn(fmt.Sprintf(
		"GIF resolution is limited to %d and dimension size is limited to %d. The image is rescaled to %dx%d",
		int(gifMaxResolution), int(gifMaxDimension), img.Width(), img.Height(),
	))

	return nil
}

func fixIcoSize(img *vips.Image) error {
	icoLimitShrink := float64(max(img.Width(), img.Height())) / icoMaxDimension

	if icoLimitShrink <= 1.0 {
		return nil
	}

	scale := 1.0 / icoLimitShrink
	if err := img.Resize(scale, scale); err != nil {
		return err
	}

	slog.Warn(fmt.Sprintf(
		"ICO dimension size is limited to %d. The image is rescaled to %dx%d",
		int(icoMaxDimension), img.Width(), img.Height(),
	))

	return nil
}

func fixSize(c *Context) error {
	switch c.PO.Format {
	case imagetype.WEBP:
		return fixWebpSize(c.Img)
	case imagetype.AVIF, imagetype.HEIC:
		return fixHeifSize(c.Img)
	case imagetype.GIF:
		return fixGifSize(c.Img)
	case imagetype.ICO:
		return fixIcoSize(c.Img)
	}

	return nil
}
