package options

import (
	"encoding/base64"
	"slices"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

func applyWidthOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.Width, args...)
}

func applyHeightOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.Height, args...)
}

func applyMinWidthOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.MinWidth, args...)
}

func applyMinHeightOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.MinHeight, args...)
}

func applyEnlargeOption(o Options, args []string) error {
	return parseBool(o, keys.Enlarge, args...)
}

func applyExtendOption(o Options, args []string) error {
	return parseExtend(o, keys.PrefixExtend, args)
}

func applyExtendAspectRatioOption(o Options, args []string) error {
	return parseExtend(o, keys.PrefixExtendAspectRatio, args)
}

func applySizeOption(o Options, args []string) (err error) {
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

func applyResizingTypeOption(o Options, args []string) error {
	if err := ensureMaxArgs(keys.ResizingType, args, 1); err != nil {
		return err
	}

	if r, ok := resizeTypes[args[0]]; ok {
		o[keys.ResizingType] = r
	} else {
		return newOptionArgumentError("Invalid %s: %s", keys.ResizingType, args[0])
	}

	return nil
}

func applyResizeOption(o Options, args []string) error {
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

func applyZoomOption(o Options, args []string) error {
	if err := ensureMaxArgs("zoom", args, 2); err != nil {
		return err
	}

	if err := parsePositiveNonZeroFloat(o, keys.ZoomWidth, args[0]); err != nil {
		return err
	}

	if len(args) < 2 {
		CopyValue(o, keys.ZoomWidth, keys.ZoomHeight)
		return nil
	}

	if err := parsePositiveNonZeroFloat(o, keys.ZoomHeight, args[1]); err != nil {
		return err
	}

	return nil
}

func applyDprOption(o Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Dpr, args...)
}

func applyGravityOption(o Options, args []string) error {
	return parseGravity(o, keys.Gravity, args, cropGravityTypes)
}

func applyCropOption(o Options, args []string) error {
	if err := parsePositiveFloat(o, keys.CropWidth, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := parsePositiveFloat(o, keys.CropHeight, args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return parseGravity(o, keys.CropGravity, args[2:], cropGravityTypes)
	}

	return nil
}

func applyPaddingOption(o Options, args []string) error {
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
		CopyValue(o, keys.PaddingTop, keys.PaddingRight)
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingBottom, args[2]); err != nil {
			return err
		}
	} else {
		CopyValue(o, keys.PaddingTop, keys.PaddingBottom)
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := parsePositiveInt(o, keys.PaddingLeft, args[3]); err != nil {
			return err
		}
	} else {
		CopyValue(o, keys.PaddingRight, keys.PaddingLeft)
	}

	o[keys.PaddingEnabled] = Get(o, keys.PaddingTop, 0) != 0 ||
		Get(o, keys.PaddingRight, 0) != 0 ||
		Get(o, keys.PaddingBottom, 0) != 0 ||
		Get(o, keys.PaddingLeft, 0) != 0

	return nil
}

func applyTrimOption(o Options, args []string) error {
	if err := ensureMaxArgs("trim", args, 4); err != nil {
		return err
	}

	nArgs := len(args)

	if err := parseFloat(o, keys.TrimThreshold, args[0]); err != nil {
		return err
	}

	o[keys.TrimEnabled] = true

	if nArgs > 1 && len(args[1]) > 0 {
		if c, err := vips.ColorFromHex(args[1]); err == nil {
			o[keys.TrimColor] = c
			o[keys.TrimSmart] = false
		} else {
			return newOptionArgumentError("Invalid %s: %s", keys.TrimColor, args[1])
		}
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

func applyRotateOption(o Options, args []string) error {
	if err := parseInt(o, keys.Rotate, args...); err != nil {
		return err
	}

	if Get(o, keys.Rotate, 0)%90 != 0 {
		return newOptionArgumentError("Rotation angle must be a multiple of 90")
	}

	return nil
}

func applyQualityOption(o Options, args []string) error {
	return parseQualityInt(o, keys.Quality, args...)
}

func applyFormatQualityOption(o Options, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError("Missing %s for: %s", keys.FormatQuality, args[argsLen-1])
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.GetTypeByName(args[i])
		if !ok {
			return newOptionArgumentError("Invalid image format: %s", args[i])
		}

		if err := parseQualityInt(o, keys.FormatQuality+"."+f.String(), args[i+1]); err != nil {
			return err
		}
	}

	return nil
}

func applyMaxBytesOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.MaxBytes, args...)
}

func applyBackgroundOption(o Options, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			o[keys.Flatten] = false
			return nil
		}

		if c, err := vips.ColorFromHex(args[0]); err == nil {
			o[keys.Flatten] = true
			o[keys.Background] = c
		} else {
			return newOptionArgumentError("Invalid %s argument: %s", keys.Background, err)
		}

	case 3:
		var c vips.Color

		if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r <= 255 {
			c.R = uint8(r)
		} else {
			return newOptionArgumentError("Invalid %s red channel: %s", keys.Background, args[0])
		}

		if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g <= 255 {
			c.G = uint8(g)
		} else {
			return newOptionArgumentError("Invalid %s green channel: %s", keys.Background, args[1])
		}

		if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b <= 255 {
			c.B = uint8(b)
		} else {
			return newOptionArgumentError("Invalid %s blue channel: %s", keys.Background, args[2])
		}

		o[keys.Flatten] = true
		o[keys.Background] = c

	default:
		return newOptionArgumentError("Invalid %s arguments: %v", keys.Background, args)
	}

	return nil
}

func applyBlurOption(o Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Blur, args...)
}

func applySharpenOption(o Options, args []string) error {
	return parsePositiveNonZeroFloat(o, keys.Sharpen, args...)
}

func applyPixelateOption(o Options, args []string) error {
	return parsePositiveInt(o, keys.Pixelate, args...)
}

func applyWatermarkOption(o Options, args []string) error {
	if err := ensureMaxArgs("watermark", args, 7); err != nil {
		return err
	}

	if wo, err := strconv.ParseFloat(args[0], 64); err == nil && wo >= 0 && wo <= 1 {
		o[keys.WatermarkEnabled] = wo > 0
		o[keys.WatermarkOpacity] = wo
	} else {
		return newOptionArgumentError("Invalid %s: %s", keys.WatermarkOpacity, args[0])
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if pos, ok := gravityTypes[args[1]]; ok && slices.Contains(watermarkGravityTypes, pos) {
			o[keys.WatermarkPosition] = pos
		} else {
			return newOptionArgumentError("Invalid %s: %s", keys.WatermarkPosition, args[1])
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

func applyFormatOption(o Options, args []string) error {
	if err := ensureMaxArgs(keys.Format, args, 1); err != nil {
		return err
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		o[keys.Format] = f
	} else {
		return newOptionArgumentError("Invalid image format: %s", args[0])
	}

	return nil
}

func applyCacheBusterOption(o Options, args []string) error {
	if err := ensureMaxArgs(keys.CacheBuster, args, 1); err != nil {
		return err
	}

	o[keys.CacheBuster] = args[0]

	return nil
}

func applySkipProcessingFormatsOption(o Options, args []string) error {
	for _, format := range args {
		if f, ok := imagetype.GetTypeByName(format); ok {
			Append(o, keys.SkipProcessing, f)
		} else {
			return newOptionArgumentError("Invalid image format in %s: %s", keys.SkipProcessing, format)
		}
	}

	return nil
}

func applyRawOption(o Options, args []string) error {
	return parseBool(o, keys.Raw, args...)
}

func applyFilenameOption(o Options, args []string) error {
	if err := ensureMaxArgs(keys.Filename, args, 2); err != nil {
		return err
	}

	filename := args[0]

	if len(args) > 1 && len(args[1]) > 0 {
		if encoded, _ := strconv.ParseBool(args[1]); encoded {
			if decoded, err := base64.RawURLEncoding.DecodeString(filename); err == nil {
				filename = string(decoded)
			} else {
				return newOptionArgumentError("Invalid %s encoding: %s", keys.Filename, err)
			}
		}
	}

	o[keys.Filename] = filename

	return nil
}

func applyExpiresOption(o Options, args []string) error {
	if err := ensureMaxArgs(keys.Expires, args, 1); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return newOptionArgumentError("Invalid %s argument: %v", keys.Expires, args[0])
	}

	if timestamp > 0 && timestamp < time.Now().Unix() {
		return newOptionArgumentError("Expired URL")
	}

	o[keys.Expires] = time.Unix(timestamp, 0)

	return nil
}

func applyStripMetadataOption(o Options, args []string) error {
	return parseBool(o, keys.StripMetadata, args...)
}

func applyKeepCopyrightOption(o Options, args []string) error {
	return parseBool(o, keys.KeepCopyright, args...)
}

func applyStripColorProfileOption(o Options, args []string) error {
	return parseBool(o, keys.StripColorProfile, args...)
}

func applyAutoRotateOption(o Options, args []string) error {
	return parseBool(o, keys.AutoRotate, args...)
}

func applyEnforceThumbnailOption(o Options, args []string) error {
	return parseBool(o, keys.EnforceThumbnail, args...)
}

func applyReturnAttachmentOption(o Options, args []string) error {
	return parseBool(o, keys.ReturnAttachment, args...)
}

func applyMaxSrcResolutionOption(f *Factory, o Options, args []string) error {
	if err := f.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseResolution(o, keys.MaxSrcResolution, args...)
}

func applyMaxSrcFileSizeOption(f *Factory, o Options, args []string) error {
	if err := f.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(o, keys.MaxSrcFileSize, args...)
}

func applyMaxAnimationFramesOption(f *Factory, o Options, args []string) error {
	if err := f.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parsePositiveNonZeroInt(o, keys.MaxAnimationFrames, args...)
}

func applyMaxAnimationFrameResolutionOption(f *Factory, o Options, args []string) error {
	if err := f.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseResolution(o, keys.MaxAnimationFrameResolution, args...)
}

func applyMaxResultDimensionOption(f *Factory, o Options, args []string) error {
	if err := f.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(o, keys.MaxResultDimension, args...)
}

func applyPresetOption(f *Factory, o Options, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if p, ok := f.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				log.Warningf("Recursive preset usage is detected: %s", preset)
				continue
			}

			Append(o, keys.UsedPresets, preset)

			if err := f.applyURLOptions(o, p, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError("Unknown preset: %s", preset)
		}
	}

	return nil
}
