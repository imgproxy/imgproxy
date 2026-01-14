package processing

import (
	"github.com/imgproxy/imgproxy/v3/vips"
)

func (p *Processor) colorspaceToProcessing(c *Context) error {
	if err := c.Img.Rad2Float(); err != nil {
		return err
	}

	supportsHDR := c.PO.Format().SupportsHDR() && p.config.PreserveHDR
	cs := guessTargetColorspace(c.Img, supportsHDR)

	if c.Img.IsLinear() {
		// If we keep its ICC, we'll get wrong colors after converting it to target
		// colorspace (we never convert back to linear).
		c.Img.RemoveColourProfile()
	} else {
		// vips 8.15+ tends to lose the colour profile during some color conversions.
		// We need to backup the colour profile before the conversion and restore it later.
		c.Img.BackupColourProfile()

		if err := c.Img.ImportColourProfile(); err != nil {
			return err
		}
	}

	// Convert to processing colorspace
	return c.Img.Colorspace(cs)
}

// guessTargetColorspace returns the colorspace to which the image should be saved.
// If target format supports 16-bit colorspace, it will be preferred.
func guessTargetColorspace(img *vips.Image, supports16Bit bool) vips.Interpretation {
	format := img.GuessInterpretation()

	switch format {
	case vips.InterpretationRGB, vips.InterpretationSRGB, vips.InterpretationBW: // 3 bytes
		return format // as is

	case vips.InterpretationRGB16: // 3 uint16
		if supports16Bit {
			return format // as is
		}
		return vips.InterpretationSRGB

	case vips.InterpretationGrey16: // 1 uint16
		if supports16Bit {
			return format // as is
		}
		return vips.InterpretationBW

	case vips.InterpretationCMYK: // 4 bytes
		return vips.InterpretationSRGB

	default:
		if supports16Bit {
			return vips.InterpretationRGB16 // best effort
		}
		return vips.InterpretationSRGB // sRGB can be produced from any colorspace
	}
}
