package vips

/*
#include "vips.h"
*/
import "C"

type Interpretation C.VipsInterpretation

const (
	InterpretationMultiBand = C.VIPS_INTERPRETATION_MULTIBAND
	InterpretationBW        = C.VIPS_INTERPRETATION_B_W
	InterpretationCMYK      = C.VIPS_INTERPRETATION_CMYK
	InterpretationRGB       = C.VIPS_INTERPRETATION_RGB
	InterpretationSRGB      = C.VIPS_INTERPRETATION_sRGB
	InterpretationRGB16     = C.VIPS_INTERPRETATION_RGB16
	InterpretationGrey16    = C.VIPS_INTERPRETATION_GREY16
	InterpretationScRGB     = C.VIPS_INTERPRETATION_scRGB
)
