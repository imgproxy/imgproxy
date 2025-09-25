package options

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips/color"
)

func applyWidthOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.Width, args...)
}

func applyHeightOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.Height, args...)
}

func applyMinWidthOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.MinWidth, args...)
}

func applyMinHeightOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.MinHeight, args...)
}

func applyEnlargeOption(o *Options, args []string) error {
	return parseBool(o, keys.Enlarge, args...)
}

func applyExtendOption(o *Options, args []string) error {
	return parseExtend(o, keys.PrefixExtend, args)
}

func applyExtendAspectRatioOption(o *Options, args []string) error {
	return parseExtend(o, keys.PrefixExtendAspectRatio, args)
}

func applySizeOption(o *Options, args []string) (err error) {
	if err = ensureMaxArgs("size", args, 7); err != nil {
		return
	}

	if len(args) >= 1 && len(args[0]) > 0 {
		if err = applyWidthOption(o, args[0:1]); err != nil {
			return
		}
	}

	if len(args) >= 2 && len(args[1]) > 0 {
		if err = applyHeightOption(o, args[1:2]); err != nil {
			return
		}
	}

	if len(args) >= 3 && len(args[2]) > 0 {
		if err = applyEnlargeOption(o, args[2:3]); err != nil {
			return
		}
	}

	if len(args) >= 4 && len(args[3]) > 0 {
		if err = applyExtendOption(o, args[3:]); err != nil {
			return
		}
	}

	return nil
}

func applyResizingTypeOption(o *Options, args []string) error {
	if err := ensureMaxArgs(keys.ResizingType, args, 1); err != nil {
		return err
	}

	if r, ok := resizeTypes[args[0]]; ok {
		o.Set(keys.ResizingType, r)
	} else {
		return newInvalidArgumentError(
			keys.ResizingType, args[0], slices.Collect(maps.Keys(resizeTypes))...,
		)
	}

	return nil
}

func applyResizeOption(o *Options, args []string) error {
	if err := ensureMaxArgs("resize", args, 8); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := applyResizingTypeOption(o, args[0:1]); err != nil {
			return err
		}
	}

	if len(args) > 1 {
		if err := applySizeOption(o, args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func applyZoomOption(o *Options, args []string) error {
	if err := ensureMaxArgs("zoom", args, 2); err != nil {
		return err
	}

	if err := parsePositiveNonZeroFloat(o, keys.ZoomWidth, args[0]); err != nil {
		return err
	}

	if len(args) < 2 {
		o.CopyValue(keys.ZoomWidth, keys.ZoomHeight)
		return nil
	}

	if err := parsePositiveNonZeroFloat(o, keys.ZoomHeight, args[1]); err != nil {
		return err
	}

	return nil
}

func applyDprOption(o *Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Dpr, args...)
}

func applyGravityOption(o *Options, args []string) error {
	return parseGravity(o, keys.Gravity, cropGravityTypes, args...)
}

func applyCropOption(o *Options, args []string) error {
	if err := parsePositiveFloat(o, keys.CropWidth, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := parsePositiveFloat(o, keys.CropHeight, args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return parseGravity(o, keys.CropGravity, cropGravityTypes, args[2:]...)
	}

	return nil
}

func applyPaddingOption(o *Options, args []string) error {
	if err := ensureMaxArgs("padding", args, 4); err != nil {
		return err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingTop, args[0]); err != nil {
			return err
		}
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingRight, args[1]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingRight)
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingBottom, args[2]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingBottom)
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingLeft, args[3]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingRight, keys.PaddingLeft)
	}

	return nil
}

func applyTrimOption(o *Options, args []string) error {
	if err := ensureMaxArgs("trim", args, 4); err != nil {
		return err
	}

	nArgs := len(args)

	if len(args[0]) > 0 {
		if err := parseFloat(o, keys.TrimThreshold, args[0]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimThreshold)
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if err := parseHexRGBColor(o, keys.TrimColor, args[1]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimColor)
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := parseBool(o, keys.TrimEqualHor, args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := parseBool(o, keys.TrimEqualVer, args[3]); err != nil {
			return err
		}
	}

	return nil
}

func applyRotateOption(o *Options, args []string) error {
	if err := parseInt(o, keys.Rotate, args...); err != nil {
		return err
	}

	if Get(o, keys.Rotate, 0)%90 != 0 {
		return newOptionArgumentError("Rotation angle must be a multiple of 90")
	}

	return nil
}

func applyQualityOption(o *Options, args []string) error {
	return parseQualityInt(o, keys.Quality, args...)
}

func applyFormatQualityOption(o *Options, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError("Missing %s for: %s", keys.PrefixFormatQuality, args[argsLen-1])
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.GetTypeByName(args[i])
		if !ok {
			return newOptionArgumentError("Invalid image format: %s", args[i])
		}

		if err := parseQualityInt(o, keys.FormatQuality(f), args[i+1]); err != nil {
			return err
		}
	}

	return nil
}

func applyMaxBytesOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.MaxBytes, args...)
}

func applyBackgroundOption(o *Options, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			o.Delete(keys.Background)
			return nil
		}

		if err := parseHexRGBColor(o, keys.Background, args[0]); err != nil {
			return err
		}

	case 3:
		var c color.RGB

		if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r <= 255 {
			c.R = uint8(r)
		} else {
			return newInvalidArgumentError(keys.Background+".R", args[0], "number in range 0-255")
		}

		if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g <= 255 {
			c.G = uint8(g)
		} else {
			return newInvalidArgumentError(keys.Background+".G", args[1], "number in range 0-255")
		}

		if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b <= 255 {
			c.B = uint8(b)
		} else {
			return newInvalidArgumentError(keys.Background+".B", args[2], "number in range 0-255")
		}

		o.Set(keys.Background, c)

	default:
		return newInvalidArgsError(keys.Background, args)
	}

	return nil
}

func applyBlurOption(o *Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Blur, args...)
}

func applySharpenOption(o *Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Sharpen, args...)
}

func applyPixelateOption(o *Options, args []string) error {
	return parsePositiveInt(o, keys.Pixelate, args...)
}

func applyWatermarkOption(o *Options, args []string) error {
	if err := ensureMaxArgs("watermark", args, 7); err != nil {
		return err
	}

	if err := parseOpacityFloat(o, keys.WatermarkOpacity, args[0]); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if _, err := parseGravityType(
			o, keys.WatermarkPosition, watermarkGravityTypes, args[1],
		); err != nil {
			return err
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := parseFloat(o, keys.WatermarkXOffset, args[2]); err != nil {
			return err
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := parseFloat(o, keys.WatermarkYOffset, args[3]); err != nil {
			return err
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if err := parsePositiveNonZeroFloat(o, keys.WatermarkScale, args[4]); err == nil {
			return err
		}
	}

	return nil
}

func applyFormatOption(o *Options, args []string) error {
	if err := ensureMaxArgs(keys.Format, args, 1); err != nil {
		return err
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		o.Set(keys.Format, f)
	} else {
		return newInvalidArgumentError(keys.Format, args[0], "supported image format")
	}

	return nil
}

func applyCacheBusterOption(o *Options, args []string) error {
	if err := ensureMaxArgs(keys.CacheBuster, args, 1); err != nil {
		return err
	}

	o.Set(keys.CacheBuster, args[0])

	return nil
}

func applySkipProcessingFormatsOption(o *Options, args []string) error {
	for _, format := range args {
		if f, ok := imagetype.GetTypeByName(format); ok {
			AppendToSlice(o, keys.SkipProcessing, f)
		} else {
			return newOptionArgumentError("Invalid image format in %s: %s", keys.SkipProcessing, format)
		}
	}

	return nil
}

func applyRawOption(o *Options, args []string) error {
	return parseBool(o, keys.Raw, args...)
}

func applyFilenameOption(o *Options, args []string) error {
	if err := ensureMaxArgs(keys.Filename, args, 2); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if encoded, _ := strconv.ParseBool(args[1]); encoded {
			return parseBase64String(o, keys.Filename, args[0])
		}
	}

	o.Set(keys.Filename, args[0])

	return nil
}

func applyExpiresOption(o *Options, args []string) error {
	if err := ensureMaxArgs(keys.Expires, args, 1); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return newInvalidArgumentError(keys.Expires, args[0], "unix timestamp")
	}

	if timestamp > 0 && timestamp < time.Now().Unix() {
		return newOptionArgumentError("Expired URL")
	}

	o.Set(keys.Expires, time.Unix(timestamp, 0))

	return nil
}

func applyStripMetadataOption(o *Options, args []string) error {
	return parseBool(o, keys.StripMetadata, args...)
}

func applyKeepCopyrightOption(o *Options, args []string) error {
	return parseBool(o, keys.KeepCopyright, args...)
}

func applyStripColorProfileOption(o *Options, args []string) error {
	return parseBool(o, keys.StripColorProfile, args...)
}

func applyAutoRotateOption(o *Options, args []string) error {
	return parseBool(o, keys.AutoRotate, args...)
}

func applyEnforceThumbnailOption(o *Options, args []string) error {
	return parseBool(o, keys.EnforceThumbnail, args...)
}

func applyReturnAttachmentOption(o *Options, args []string) error {
	return parseBool(o, keys.ReturnAttachment, args...)
}

func applyMaxSrcResolutionOption(p *Parser, o *Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseResolution(o, keys.MaxSrcResolution, args...)
}

func applyMaxSrcFileSizeOption(p *Parser, o *Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(o, keys.MaxSrcFileSize, args...)
}

func applyMaxAnimationFramesOption(p *Parser, o *Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parsePositiveNonZeroInt(o, keys.MaxAnimationFrames, args...)
}

func applyMaxAnimationFrameResolutionOption(p *Parser, o *Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseResolution(o, keys.MaxAnimationFrameResolution, args...)
}

func applyMaxResultDimensionOption(p *Parser, o *Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(o, keys.MaxResultDimension, args...)
}

func applyPresetOption(p *Parser, o *Options, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if pr, ok := p.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				slog.Warn(fmt.Sprintf("Recursive preset usage is detected: %s", preset))
				continue
			}

			AppendToSlice(o, keys.UsedPresets, preset)

			if err := p.applyURLOptions(o, pr, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError("Unknown preset: %s", preset)
		}
	}

	return nil
}
