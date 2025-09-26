package vips

/*
#include "options.h"
*/
import "C"
import "github.com/imgproxy/imgproxy/v3/options"

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
