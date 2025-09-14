package options

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func applyWidthOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.Width, "width", args...)
}

func applyHeightOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.Height, "height", args...)
}

func applyMinWidthOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.MinWidth, "min width", args...)
}

func applyMinHeightOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.MinHeight, "min height", args...)
}

func applyEnlargeOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.Enlarge, "enlarge", args...)
}

func applyExtendOption(po *ProcessingOptions, args []string) error {
	return parseExtend(&po.Extend, "extend", args)
}

func applyExtendAspectRatioOption(po *ProcessingOptions, args []string) error {
	return parseExtend(&po.ExtendAspectRatio, "extend_aspect_ratio", args)
}

func applySizeOption(po *ProcessingOptions, args []string) (err error) {
	if err = ensureMaxArgs("size", args, 7); err != nil {
		return
	}

	if len(args) >= 1 && len(args[0]) > 0 {
		if err = applyWidthOption(po, args[0:1]); err != nil {
			return
		}
	}

	if len(args) >= 2 && len(args[1]) > 0 {
		if err = applyHeightOption(po, args[1:2]); err != nil {
			return
		}
	}

	if len(args) >= 3 && len(args[2]) > 0 {
		if err = applyEnlargeOption(po, args[2:3]); err != nil {
			return
		}
	}

	if len(args) >= 4 && len(args[3]) > 0 {
		if err = applyExtendOption(po, args[3:]); err != nil {
			return
		}
	}

	return nil
}

func applyResizingTypeOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("resizing type", args, 1); err != nil {
		return err
	}

	if r, ok := resizeTypes[args[0]]; ok {
		po.ResizingType = r
	} else {
		return newOptionArgumentError("Invalid resize type: %s", args[0])
	}

	return nil
}

func applyResizeOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("resize", args, 8); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := applyResizingTypeOption(po, args[0:1]); err != nil {
			return err
		}
	}

	if len(args) > 1 {
		if err := applySizeOption(po, args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func applyZoomOption(po *ProcessingOptions, args []string) error {
	nArgs := len(args)

	if err := ensureMaxArgs("zoom", args, 2); err != nil {
		return err
	}

	var z float64
	if err := parsePositiveNonZeroFloat64(&z, "zoom", args[0]); err != nil {
		return err
	}

	po.ZoomWidth = z
	po.ZoomHeight = z

	if nArgs > 1 {
		if err := parsePositiveNonZeroFloat64(&po.ZoomHeight, "zoom height", args[1]); err != nil {
			return err
		}
	}

	return nil
}

func applyDprOption(po *ProcessingOptions, args []string) error {
	return parsePositiveNonZeroFloat64(&po.Dpr, "dpr", args...)
}

func applyGravityOption(po *ProcessingOptions, args []string) error {
	return parseGravity(&po.Gravity, "gravity", args, cropGravityTypes)
}

func applyCropOption(po *ProcessingOptions, args []string) error {
	if err := parsePositiveFloat64(&po.Crop.Width, "crop width", args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := parsePositiveFloat64(&po.Crop.Height, "crop height", args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return parseGravity(&po.Crop.Gravity, "crop gravity", args[2:], cropGravityTypes)
	}

	return nil
}

func applyPaddingOption(po *ProcessingOptions, args []string) error {
	nArgs := len(args)

	if nArgs < 1 || nArgs > 4 {
		return newOptionArgumentError("Invalid padding arguments: %v", args)
	}

	po.Padding.Enabled = true

	if nArgs > 0 && len(args[0]) > 0 {
		if err := parsePositiveInt(&po.Padding.Top, "padding top (+all)", args[0]); err != nil {
			return err
		}
		po.Padding.Right = po.Padding.Top
		po.Padding.Bottom = po.Padding.Top
		po.Padding.Left = po.Padding.Top
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if err := parsePositiveInt(&po.Padding.Right, "padding right (+left)", args[1]); err != nil {
			return err
		}
		po.Padding.Left = po.Padding.Right
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := parsePositiveInt(&po.Padding.Bottom, "padding bottom", args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := parsePositiveInt(&po.Padding.Left, "padding left", args[3]); err != nil {
			return err
		}
	}

	if po.Padding.Top == 0 && po.Padding.Right == 0 && po.Padding.Bottom == 0 && po.Padding.Left == 0 {
		po.Padding.Enabled = false
	}

	return nil
}

func applyTrimOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("trim", args, 4); err != nil {
		return err
	}

	nArgs := len(args)

	if err := parseFloat64(&po.Trim.Threshold, "trim threshold", args[0]); err != nil {
		return err
	}

	po.Trim.Enabled = true

	if nArgs > 1 && len(args[1]) > 0 {
		if c, err := vips.ColorFromHex(args[1]); err == nil {
			po.Trim.Color = c
			po.Trim.Smart = false
		} else {
			return newOptionArgumentError("Invalid trim color: %s", args[1])
		}
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := parseBool(&po.Trim.EqualHor, "trim equal horizontal", args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := parseBool(&po.Trim.EqualVer, "trim equal vertical", args[3]); err != nil {
			return err
		}
	}

	return nil
}

func applyRotateOption(po *ProcessingOptions, args []string) error {
	if err := parseInt(&po.Rotate, "rotate", args...); err != nil {
		return err
	}

	if po.Rotate%90 != 0 {
		return newOptionArgumentError("Rotation angle must be a multiple of 90")
	}

	return nil
}

func applyQualityOption(po *ProcessingOptions, args []string) error {
	return parseQualityInt(&po.Quality, "quality", args...)
}

func applyFormatQualityOption(po *ProcessingOptions, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError("Missing quality for: %s", args[argsLen-1])
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.GetTypeByName(args[i])
		if !ok {
			return newOptionArgumentError("Invalid image format: %s", args[i])
		}

		var q int
		if err := parseQualityInt(&q, args[i]+" quality", args[i+1]); err != nil {
			return err
		}

		po.FormatQuality[f] = q
	}

	return nil
}

func applyMaxBytesOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.MaxBytes, "max_bytes", args...)
}

func applyBackgroundOption(po *ProcessingOptions, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			po.Flatten = false
		} else if c, err := vips.ColorFromHex(args[0]); err == nil {
			po.Flatten = true
			po.Background = c
		} else {
			return newOptionArgumentError("Invalid background argument: %s", err)
		}

	case 3:
		po.Flatten = true

		if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r <= 255 {
			po.Background.R = uint8(r)
		} else {
			return newOptionArgumentError("Invalid background red channel: %s", args[0])
		}

		if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g <= 255 {
			po.Background.G = uint8(g)
		} else {
			return newOptionArgumentError("Invalid background green channel: %s", args[1])
		}

		if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b <= 255 {
			po.Background.B = uint8(b)
		} else {
			return newOptionArgumentError("Invalid background blue channel: %s", args[2])
		}

	default:
		return newOptionArgumentError("Invalid background arguments: %v", args)
	}

	return nil
}

func applyBlurOption(po *ProcessingOptions, args []string) error {
	return parsePositiveNonZeroFloat32(&po.Blur, "blur", args...)
}

func applySharpenOption(po *ProcessingOptions, args []string) error {
	return parsePositiveNonZeroFloat32(&po.Sharpen, "sharpen", args...)
}

func applyPixelateOption(po *ProcessingOptions, args []string) error {
	return parsePositiveInt(&po.Pixelate, "pixelate", args...)
}

func applyWatermarkOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("watermark", args, 7); err != nil {
		return err
	}

	if o, err := strconv.ParseFloat(args[0], 64); err == nil && o >= 0 && o <= 1 {
		po.Watermark.Enabled = o > 0
		po.Watermark.Opacity = o
	} else {
		return newOptionArgumentError("Invalid watermark opacity: %s", args[0])
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if g, ok := gravityTypes[args[1]]; ok && slices.Contains(watermarkGravityTypes, g) {
			po.Watermark.Position.Type = g
		} else {
			return newOptionArgumentError("Invalid watermark position: %s", args[1])
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := parseFloat64(&po.Watermark.Position.X, "watermark X offset", args[2]); err != nil {
			return err
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := parseFloat64(&po.Watermark.Position.Y, "watermark Y offset", args[3]); err != nil {
			return err
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if err := parsePositiveNonZeroFloat64(&po.Watermark.Scale, "watermark scale", args[4]); err == nil {
			return err
		}
	}

	return nil
}

func applyFormatOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("format", args, 1); err != nil {
		return err
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		po.Format = f
	} else {
		return newOptionArgumentError("Invalid image format: %s", args[0])
	}

	return nil
}

func applyCacheBusterOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("cache buster", args, 1); err != nil {
		return err
	}

	po.CacheBuster = args[0]

	return nil
}

func applySkipProcessingFormatsOption(po *ProcessingOptions, args []string) error {
	for _, format := range args {
		if f, ok := imagetype.GetTypeByName(format); ok {
			po.SkipProcessingFormats = append(po.SkipProcessingFormats, f)
		} else {
			return newOptionArgumentError("Invalid image format in skip processing: %s", format)
		}
	}

	return nil
}

func applyRawOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.Raw, "raw", args...)
}

func applyFilenameOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("filename", args, 2); err != nil {
		return err
	}

	po.Filename = args[0]

	if len(args) == 1 {
		return nil
	}

	var b bool
	if err := parseBool(&b, "filename is base64", args[1]); err != nil || !b {
		return err
	}

	decoded, err := base64.RawURLEncoding.DecodeString(po.Filename)
	if err != nil {
		return newOptionArgumentError("Invalid filename encoding: %s", err)
	}

	po.Filename = string(decoded)

	return nil
}

func applyExpiresOption(po *ProcessingOptions, args []string) error {
	if err := ensureMaxArgs("expires", args, 1); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return newOptionArgumentError("Invalid expires argument: %v", args[0])
	}

	if timestamp > 0 && timestamp < time.Now().Unix() {
		return newOptionArgumentError("Expired URL")
	}

	expires := time.Unix(timestamp, 0)
	po.Expires = &expires

	return nil
}

func applyStripMetadataOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.StripMetadata, "strip metadata", args...)
}

func applyKeepCopyrightOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.KeepCopyright, "keep copyright", args...)
}

func applyStripColorProfileOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.StripColorProfile, "strip color profile", args...)
}

func applyAutoRotateOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.AutoRotate, "auto rotate", args...)
}

func applyEnforceThumbnailOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.EnforceThumbnail, "enforce thumbnail", args...)
}

func applyReturnAttachmentOption(po *ProcessingOptions, args []string) error {
	return parseBool(&po.ReturnAttachment, "return_attachment", args...)
}

func applyMaxSrcResolutionOption(po *ProcessingOptions, args []string) error {
	if err := po.isSecurityOptionsAllowed(); err != nil {
		return err
	}

	var v float64
	if err := parsePositiveNonZeroFloat64(&v, "max_src_resolution", args...); err != nil {
		return err
	}

	po.SecurityOptions.MaxSrcResolution = int(v * 1000000)

	return nil
}

func applyMaxSrcFileSizeOption(po *ProcessingOptions, args []string) error {
	if err := po.isSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(&po.SecurityOptions.MaxSrcFileSize, "max_src_file_size", args...)
}

func applyMaxAnimationFramesOption(po *ProcessingOptions, args []string) error {
	if err := po.isSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parsePositiveNonZeroInt(&po.SecurityOptions.MaxAnimationFrames, "max_animation_frames", args...)
}

func applyMaxAnimationFrameResolutionOption(po *ProcessingOptions, args []string) error {
	if err := po.isSecurityOptionsAllowed(); err != nil {
		return err
	}

	var v float64
	if err := parseFloat64(&v, "max_animation_frame_resolution", args...); err != nil {
		return err
	}

	po.SecurityOptions.MaxAnimationFrameResolution = int(v * 1000000)

	return nil
}

func applyMaxResultDimensionOption(po *ProcessingOptions, args []string) error {
	if err := po.isSecurityOptionsAllowed(); err != nil {
		return err
	}

	return parseInt(&po.SecurityOptions.MaxResultDimension, "max_result_dimension", args...)
}

func applyPresetOption(f *Factory, po *ProcessingOptions, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if p, ok := f.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				slog.Warn(fmt.Sprintf("Recursive preset usage is detected: %s", preset))
				continue
			}

			po.UsedPresets = append(po.UsedPresets, preset)

			if err := f.applyURLOptions(po, p, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError("Unknown preset: %s", preset)
		}
	}

	return nil
}
