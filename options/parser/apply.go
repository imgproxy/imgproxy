package optionsparser

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/vips/color"
)

func (p *Parser) applyWidthOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.Width, args...)
}

func (p *Parser) applyHeightOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.Height, args...)
}

func (p *Parser) applyMinWidthOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.MinWidth, args...)
}

func (p *Parser) applyMinHeightOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.MinHeight, args...)
}

func (p *Parser) applyEnlargeOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.Enlarge, args...)
}

func (p *Parser) applyExtendOption(o *options.Options, args []string) error {
	return p.parseExtend(o, keys.PrefixExtend, args)
}

func (p *Parser) applyExtendAspectRatioOption(o *options.Options, args []string) error {
	return p.parseExtend(o, keys.PrefixExtendAspectRatio, args)
}

func (p *Parser) applySizeOption(o *options.Options, args []string) (err error) {
	if err = p.ensureMaxArgs("size", args, 7); err != nil {
		return
	}

	if len(args) >= 1 && len(args[0]) > 0 {
		if err = p.applyWidthOption(o, args[0:1]); err != nil {
			return
		}
	}

	if len(args) >= 2 && len(args[1]) > 0 {
		if err = p.applyHeightOption(o, args[1:2]); err != nil {
			return
		}
	}

	if len(args) >= 3 && len(args[2]) > 0 {
		if err = p.applyEnlargeOption(o, args[2:3]); err != nil {
			return
		}
	}

	if len(args) >= 4 && len(args[3]) > 0 {
		if err = p.applyExtendOption(o, args[3:]); err != nil {
			return
		}
	}

	return nil
}

func (p *Parser) applyResizingTypeOption(o *options.Options, args []string) error {
	return parseFromMap(p, o, keys.ResizingType, processing.ResizeTypes, args...)
}

func (p *Parser) applyResizeOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("resize", args, 8); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := p.applyResizingTypeOption(o, args[0:1]); err != nil {
			return err
		}
	}

	if len(args) > 1 {
		if err := p.applySizeOption(o, args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyZoomOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("zoom", args, 2); err != nil {
		return err
	}

	if err := p.parsePositiveNonZeroFloat(o, keys.ZoomWidth, args[0]); err != nil {
		return err
	}

	if len(args) < 2 {
		o.CopyValue(keys.ZoomWidth, keys.ZoomHeight)
		return nil
	}

	if err := p.parsePositiveNonZeroFloat(o, keys.ZoomHeight, args[1]); err != nil {
		return err
	}

	return nil
}

func (p *Parser) applyDprOption(o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(o, keys.Dpr, args...)
}

func (p *Parser) applyGravityOption(o *options.Options, args []string) error {
	return p.parseGravity(o, keys.Gravity, processing.CropGravityTypes, args...)
}

func (p *Parser) applyCropOption(o *options.Options, args []string) error {
	if err := p.parsePositiveFloat(o, keys.CropWidth, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := p.parsePositiveFloat(o, keys.CropHeight, args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return p.parseGravity(o, keys.CropGravity, processing.CropGravityTypes, args[2:]...)
	}

	return nil
}

func (p *Parser) applyPaddingOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("padding", args, 4); err != nil {
		return err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if err := p.parsePositiveInt(o, keys.PaddingTop, args[0]); err != nil {
			return err
		}
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if err := p.parsePositiveInt(o, keys.PaddingRight, args[1]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingRight)
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := p.parsePositiveInt(o, keys.PaddingBottom, args[2]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingBottom)
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := p.parsePositiveInt(o, keys.PaddingLeft, args[3]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingRight, keys.PaddingLeft)
	}

	return nil
}

func (p *Parser) applyTrimOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("trim", args, 4); err != nil {
		return err
	}

	nArgs := len(args)

	if len(args[0]) > 0 {
		if err := p.parseFloat(o, keys.TrimThreshold, args[0]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimThreshold)
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if err := p.parseHexRGBColor(o, keys.TrimColor, args[1]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimColor)
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := p.parseBool(o, keys.TrimEqualHor, args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := p.parseBool(o, keys.TrimEqualVer, args[3]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyRotateOption(o *options.Options, args []string) error {
	if err := p.parseInt(o, keys.Rotate, args...); err != nil {
		return err
	}

	if options.Get(o, keys.Rotate, 0)%90 != 0 {
		return newOptionArgumentError(keys.Rotate, "Rotation angle must be a multiple of 90")
	}

	return nil
}

func (p *Parser) applyFlipOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("flip", args, 2); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := p.parseBool(o, keys.FlipHorizontal, args[0]); err != nil {
			return err
		}
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if err := p.parseBool(o, keys.FlipVertical, args[1]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyQualityOption(o *options.Options, args []string) error {
	return p.parseQualityInt(o, keys.Quality, args...)
}

func (p *Parser) applyFormatQualityOption(o *options.Options, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError(
			keys.PrefixFormatQuality,
			"Missing %s for: %s",
			keys.PrefixFormatQuality,
			args[argsLen-1],
		)
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.GetTypeByName(args[i])
		if !ok {
			return newInvalidArgumentError(keys.PrefixFormatQuality, "Invalid image format: %s", args[i])
		}

		if err := p.parseQualityInt(o, keys.FormatQuality(f), args[i+1]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyMaxBytesOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.MaxBytes, args...)
}

func (p *Parser) applyBackgroundOption(o *options.Options, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			o.Delete(keys.Background)
			return nil
		}

		if err := p.parseHexRGBColor(o, keys.Background, args[0]); err != nil {
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

func (p *Parser) applyBlurOption(o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(o, keys.Blur, args...)
}

func (p *Parser) applySharpenOption(o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(o, keys.Sharpen, args...)
}

func (p *Parser) applyPixelateOption(o *options.Options, args []string) error {
	return p.parsePositiveInt(o, keys.Pixelate, args...)
}

func (p *Parser) applyWatermarkOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs("watermark", args, 7); err != nil {
		return err
	}

	if err := p.parseOpacityFloat(o, keys.WatermarkOpacity, args[0]); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if _, err := p.parseGravityType(
			o, keys.WatermarkPosition, processing.WatermarkGravityTypes, args[1],
		); err != nil {
			return err
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := p.parseFloat(o, keys.WatermarkXOffset, args[2]); err != nil {
			return err
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := p.parseFloat(o, keys.WatermarkYOffset, args[3]); err != nil {
			return err
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if err := p.parsePositiveNonZeroFloat(o, keys.WatermarkScale, args[4]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyFormatOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(keys.Format, args, 1); err != nil {
		return err
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		o.Set(keys.Format, f)
	} else {
		return newInvalidArgumentError(keys.Format, args[0], "supported image format")
	}

	return nil
}

func (p *Parser) applyCacheBusterOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(keys.CacheBuster, args, 1); err != nil {
		return err
	}

	if len(args[0]) == 0 {
		o.Delete(keys.CacheBuster)
		return nil
	}

	o.Set(keys.CacheBuster, args[0])

	return nil
}

func (p *Parser) applySkipProcessingFormatsOption(o *options.Options, args []string) error {
	for _, format := range args {
		if f, ok := imagetype.GetTypeByName(format); ok {
			options.AppendToSlice(o, keys.SkipProcessing, f)
		} else {
			return newOptionArgumentError(keys.SkipProcessing, "Invalid image format in %s: %s", keys.SkipProcessing, format)
		}
	}

	return nil
}

func (p *Parser) applyRawOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.Raw, args...)
}

func (p *Parser) applyFilenameOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(keys.Filename, args, 2); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if encoded, _ := strconv.ParseBool(args[1]); encoded {
			return p.parseBase64String(o, keys.Filename, args[0])
		}
	}

	o.Set(keys.Filename, args[0])

	return nil
}

func (p *Parser) applyExpiresOption(o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(keys.Expires, args, 1); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return newInvalidArgumentError(keys.Expires, args[0], "unix timestamp")
	}

	if timestamp > 0 && timestamp < time.Now().Unix() {
		return newOptionArgumentError(keys.Expires, "Expired URL")
	}

	o.Set(keys.Expires, time.Unix(timestamp, 0))

	return nil
}

func (p *Parser) applyStripMetadataOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.StripMetadata, args...)
}

func (p *Parser) applyKeepCopyrightOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.KeepCopyright, args...)
}

func (p *Parser) applyStripColorProfileOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.StripColorProfile, args...)
}

func (p *Parser) applyAutoRotateOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.AutoRotate, args...)
}

func (p *Parser) applyEnforceThumbnailOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.EnforceThumbnail, args...)
}

func (p *Parser) applyReturnAttachmentOption(o *options.Options, args []string) error {
	return p.parseBool(o, keys.ReturnAttachment, args...)
}

func (p *Parser) applyMaxSrcResolutionOption(o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return p.parseResolution(o, keys.MaxSrcResolution, args...)
}

func (p *Parser) applyMaxSrcFileSizeOption(o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return p.parseInt(o, keys.MaxSrcFileSize, args...)
}

func (p *Parser) applyMaxAnimationFramesOption(o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return p.parsePositiveNonZeroInt(o, keys.MaxAnimationFrames, args...)
}

func (p *Parser) applyMaxAnimationFrameResolutionOption(o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return p.parseResolution(o, keys.MaxAnimationFrameResolution, args...)
}

func (p *Parser) applyMaxResultDimensionOption(o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	return p.parseInt(o, keys.MaxResultDimension, args...)
}

func (p *Parser) applyPresetOption(o *options.Options, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if pr, ok := p.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				slog.Warn(fmt.Sprintf("Recursive preset usage is detected: %s", preset))
				continue
			}

			options.AppendToSlice(o, keys.UsedPresets, preset)

			if err := p.applyURLOptions(o, pr, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError("preset", "Unknown preset: %s", preset)
		}
	}

	return nil
}
