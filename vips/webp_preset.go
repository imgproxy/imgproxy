package vips

/*
#include "vips.h"
*/
import "C"

import (
	"log/slog"
	"strconv"
)

// WebpPreset represents WebP preset to use when saving WebP images
type WebpPreset C.VipsForeignWebpPreset

const (
	WebpPresetDefault = C.VIPS_FOREIGN_WEBP_PRESET_DEFAULT
	WebpPresetPhoto   = C.VIPS_FOREIGN_WEBP_PRESET_PHOTO
	WebpPresetPicture = C.VIPS_FOREIGN_WEBP_PRESET_PICTURE
	WebpPresetDrawing = C.VIPS_FOREIGN_WEBP_PRESET_DRAWING
	WebpPresetIcon    = C.VIPS_FOREIGN_WEBP_PRESET_ICON
	WebpPresetText    = C.VIPS_FOREIGN_WEBP_PRESET_TEXT
)

// WebpPresets maps string representations to WebpPreset values
var WebpPresets = map[string]WebpPreset{
	"default": WebpPresetDefault,
	"photo":   WebpPresetPhoto,
	"picture": WebpPresetPicture,
	"drawing": WebpPresetDrawing,
	"icon":    WebpPresetIcon,
	"text":    WebpPresetText,
}

// C converts WebpPreset to C.VipsForeignWebpPreset
func (wp WebpPreset) C() C.VipsForeignWebpPreset {
	return C.VipsForeignWebpPreset(wp)
}

// String returns the string representation of the WebpPreset
func (wp WebpPreset) String() string {
	for k, v := range WebpPresets {
		if v == wp {
			return k
		}
	}
	return "unknown"
}

// MarshalJSON implements the json.Marshaler interface
func (wp WebpPreset) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, wp.String()), nil
}

// LogValue implements the slog.LogValuer interface
func (wp WebpPreset) LogValue() slog.Value {
	return slog.StringValue(wp.String())
}
