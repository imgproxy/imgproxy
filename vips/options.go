package vips

/*
#include "options.h"
*/
import "C"
import (
	"github.com/imgproxy/imgproxy/v3/options"
)

func newLoadOptions(shrink float64, page, pages int) C.ImgproxyLoadOptions {
	return C.ImgproxyLoadOptions{
		Shrink:    C.double(shrink),
		Thumbnail: 0, // Don't load thumbnail by default. Set it explicitly when needed.

		Page:  C.int(page),
		Pages: C.int(pages),

		PngUnlimited: gbool(config.PngUnlimited),
		SvgUnlimited: gbool(config.SvgUnlimited),
	}
}

func newSaveOptions(_ *options.Options) C.ImgproxySaveOptions {
	return C.ImgproxySaveOptions{
		JpegProgressive: gbool(config.JpegProgressive),

		PngInterlaced:         gbool(config.PngInterlaced),
		PngQuantize:           gbool(config.PngQuantize),
		PngQuantizationColors: C.int(config.PngQuantizationColors),

		WebpPreset: C.VipsForeignWebpPreset(config.WebpPreset),
		WebpEffort: C.int(config.WebpEffort),

		AvifSpeed: C.int(config.AvifSpeed),

		JxlEffort: C.int(config.JxlEffort),
	}
}
