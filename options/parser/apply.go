package optionsparser

import (
	"context"
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

func (p *Parser) applyWidthOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.Width, args...)
}

func (p *Parser) applyHeightOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.Height, args...)
}

func (p *Parser) applyMinWidthOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.MinWidth, args...)
}

func (p *Parser) applyMinHeightOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.MinHeight, args...)
}

func (p *Parser) applyEnlargeOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.Enlarge, args...)
}

func (p *Parser) applyExtendOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseExtend(ctx, o, keys.PrefixExtend, args)
}

func (p *Parser) applyExtendAspectRatioOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseExtend(ctx, o, keys.PrefixExtendAspectRatio, args)
}

func (p *Parser) applySizeOption(ctx context.Context, o *options.Options, args []string) (err error) {
	if err = p.ensureMaxArgs(ctx, "size", args, 7); err != nil {
		return
	}

	if len(args) >= 1 && len(args[0]) > 0 {
		if err = p.applyWidthOption(ctx, o, args[0:1]); err != nil {
			return
		}
	}

	if len(args) >= 2 && len(args[1]) > 0 {
		if err = p.applyHeightOption(ctx, o, args[1:2]); err != nil {
			return
		}
	}

	if len(args) >= 3 && len(args[2]) > 0 {
		if err = p.applyEnlargeOption(ctx, o, args[2:3]); err != nil {
			return
		}
	}

	if len(args) >= 4 && len(args[3]) > 0 {
		if err = p.applyExtendOption(ctx, o, args[3:]); err != nil {
			return
		}
	}

	return nil
}

func (p *Parser) applyResizingTypeOption(ctx context.Context, o *options.Options, args []string) error {
	return parseFromMap(ctx, p, o, keys.ResizingType, processing.ResizeTypes, args...)
}

func (p *Parser) applyResizeOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "resize", args, 8); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := p.applyResizingTypeOption(ctx, o, args[0:1]); err != nil {
			return err
		}
	}

	if len(args) > 1 {
		if err := p.applySizeOption(ctx, o, args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyZoomOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "zoom", args, 2); err != nil {
		return err
	}

	if err := p.parsePositiveNonZeroFloat(ctx, o, keys.ZoomWidth, args[0]); err != nil {
		return err
	}

	if len(args) < 2 {
		o.CopyValue(keys.ZoomWidth, keys.ZoomHeight)
		return nil
	}

	if err := p.parsePositiveNonZeroFloat(ctx, o, keys.ZoomHeight, args[1]); err != nil {
		return err
	}

	return nil
}

func (p *Parser) applyDprOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(ctx, o, keys.Dpr, args...)
}

func (p *Parser) applyGravityOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseGravity(ctx, o, keys.Gravity, processing.CropGravityTypes, args...)
}

func (p *Parser) applyCropOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.parsePositiveFloat(ctx, o, keys.CropWidth, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := p.parsePositiveFloat(ctx, o, keys.CropHeight, args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return p.parseGravity(ctx, o, keys.CropGravity, processing.CropGravityTypes, args[2:]...)
	}

	return nil
}

func (p *Parser) applyPaddingOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "padding", args, 4); err != nil {
		return err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if err := p.parsePositiveInt(ctx, o, keys.PaddingTop, args[0]); err != nil {
			return err
		}
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if err := p.parsePositiveInt(ctx, o, keys.PaddingRight, args[1]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingRight)
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := p.parsePositiveInt(ctx, o, keys.PaddingBottom, args[2]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingTop, keys.PaddingBottom)
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := p.parsePositiveInt(ctx, o, keys.PaddingLeft, args[3]); err != nil {
			return err
		}
	} else {
		o.CopyValue(keys.PaddingRight, keys.PaddingLeft)
	}

	return nil
}

func (p *Parser) applyTrimOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "trim", args, 4); err != nil {
		return err
	}

	nArgs := len(args)

	if len(args[0]) > 0 {
		if err := p.parseFloat(ctx, o, keys.TrimThreshold, args[0]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimThreshold)
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if err := p.parseHexRGBColor(ctx, o, keys.TrimColor, args[1]); err != nil {
			return err
		}
	} else {
		o.Delete(keys.TrimColor)
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := p.parseBool(ctx, o, keys.TrimEqualHor, args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := p.parseBool(ctx, o, keys.TrimEqualVer, args[3]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyRotateOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.parseInt(ctx, o, keys.Rotate, args...); err != nil {
		return err
	}

	if options.Get(o, keys.Rotate, 0)%90 != 0 {
		return newOptionArgumentError(ctx, keys.Rotate, "Rotation angle must be a multiple of 90")
	}

	return nil
}

func (p *Parser) applyFlipOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "flip", args, 2); err != nil {
		return err
	}

	if len(args[0]) > 0 {
		if err := p.parseBool(ctx, o, keys.FlipHorizontal, args[0]); err != nil {
			return err
		}
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if err := p.parseBool(ctx, o, keys.FlipVertical, args[1]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyQualityOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseQualityInt(ctx, o, keys.Quality, args...)
}

func (p *Parser) applyFormatQualityOption(ctx context.Context, o *options.Options, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError(
			ctx,
			keys.PrefixFormatQuality,
			"Missing %s for: %s",
			keys.PrefixFormatQuality,
			args[argsLen-1],
		)
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.GetTypeByName(args[i])
		if !ok {
			return newInvalidArgumentError(ctx, keys.PrefixFormatQuality, "Invalid image format: %s", args[i])
		}

		if err := p.parseQualityInt(ctx, o, keys.FormatQuality(f), args[i+1]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyMaxBytesOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.MaxBytes, args...)
}

func (p *Parser) applyBackgroundOption(ctx context.Context, o *options.Options, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			o.Delete(keys.Background)
			return nil
		}

		if err := p.parseHexRGBColor(ctx, o, keys.Background, args[0]); err != nil {
			return err
		}

	case 3:
		var c color.RGB

		if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r <= 255 {
			c.R = uint8(r)
		} else {
			return newInvalidArgumentError(ctx, keys.Background+".R", args[0], "number in range 0-255")
		}

		if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g <= 255 {
			c.G = uint8(g)
		} else {
			return newInvalidArgumentError(ctx, keys.Background+".G", args[1], "number in range 0-255")
		}

		if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b <= 255 {
			c.B = uint8(b)
		} else {
			return newInvalidArgumentError(ctx, keys.Background+".B", args[2], "number in range 0-255")
		}

		o.Set(keys.Background, c)

	default:
		return newInvalidArgsError(ctx, keys.Background, args)
	}

	return nil
}

func (p *Parser) applyBlurOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(ctx, o, keys.Blur, args...)
}

func (p *Parser) applySharpenOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveNonZeroFloat(ctx, o, keys.Sharpen, args...)
}

func (p *Parser) applyPixelateOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parsePositiveInt(ctx, o, keys.Pixelate, args...)
}

func (p *Parser) applyWatermarkOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, "watermark", args, 7); err != nil {
		return err
	}

	if err := p.parseOpacityFloat(ctx, o, keys.WatermarkOpacity, args[0]); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if _, err := p.parseGravityType(
			ctx, o, keys.WatermarkPosition, processing.WatermarkGravityTypes, args[1],
		); err != nil {
			return err
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if err := p.parseFloat(ctx, o, keys.WatermarkXOffset, args[2]); err != nil {
			return err
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if err := p.parseFloat(ctx, o, keys.WatermarkYOffset, args[3]); err != nil {
			return err
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if err := p.parsePositiveNonZeroFloat(ctx, o, keys.WatermarkScale, args[4]); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) applyFormatOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, keys.Format, args, 1); err != nil {
		return err
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		o.Set(keys.Format, f)
	} else {
		return newInvalidArgumentError(ctx, keys.Format, args[0], "supported image format")
	}

	return nil
}

func (p *Parser) applyCacheBusterOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, keys.CacheBuster, args, 1); err != nil {
		return err
	}

	if len(args[0]) == 0 {
		o.Delete(keys.CacheBuster)
		return nil
	}

	o.Set(keys.CacheBuster, args[0])

	return nil
}

func (p *Parser) applySkipProcessingFormatsOption(ctx context.Context, o *options.Options, args []string) error {
	for _, format := range args {
		if f, ok := imagetype.GetTypeByName(format); ok {
			options.AppendToSlice(o, keys.SkipProcessing, f)
		} else {
			return newOptionArgumentError(
				ctx,
				keys.SkipProcessing,
				"Invalid image format in %s: %s",
				keys.SkipProcessing,
				format,
			)
		}
	}

	return nil
}

func (p *Parser) applyRawOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.Raw, args...)
}

func (p *Parser) applyFilenameOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, keys.Filename, args, 2); err != nil {
		return err
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if encoded, _ := strconv.ParseBool(args[1]); encoded {
			return p.parseBase64String(ctx, o, keys.Filename, args[0])
		}
	}

	o.Set(keys.Filename, args[0])

	return nil
}

func (p *Parser) applyExpiresOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.ensureMaxArgs(ctx, keys.Expires, args, 1); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return newInvalidArgumentError(ctx, keys.Expires, args[0], "unix timestamp")
	}

	if timestamp > 0 && timestamp < time.Now().Unix() {
		return newOptionArgumentError(ctx, keys.Expires, "Expired URL")
	}

	o.Set(keys.Expires, time.Unix(timestamp, 0))

	return nil
}

func (p *Parser) applyStripMetadataOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.StripMetadata, args...)
}

func (p *Parser) applyKeepCopyrightOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.KeepCopyright, args...)
}

func (p *Parser) applyStripColorProfileOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.StripColorProfile, args...)
}

func (p *Parser) applyAutoRotateOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.AutoRotate, args...)
}

func (p *Parser) applyEnforceThumbnailOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.EnforceThumbnail, args...)
}

func (p *Parser) applyReturnAttachmentOption(ctx context.Context, o *options.Options, args []string) error {
	return p.parseBool(ctx, o, keys.ReturnAttachment, args...)
}

func (p *Parser) applyMaxSrcResolutionOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(ctx); err != nil {
		return err
	}

	return p.parseResolution(ctx, o, keys.MaxSrcResolution, args...)
}

func (p *Parser) applyMaxSrcFileSizeOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(ctx); err != nil {
		return err
	}

	return p.parseInt(ctx, o, keys.MaxSrcFileSize, args...)
}

func (p *Parser) applyMaxAnimationFramesOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(ctx); err != nil {
		return err
	}

	return p.parsePositiveNonZeroInt(ctx, o, keys.MaxAnimationFrames, args...)
}

func (p *Parser) applyMaxAnimationFrameResolutionOption(ctx context.Context, o *options.Options, args []string) error {
	if err := p.IsSecurityOptionsAllowed(ctx); err != nil {
		return err
	}

	return p.parseResolution(ctx, o, keys.MaxAnimationFrameResolution, args...)
}

func (p *Parser) applyMaxResultDimensionOption(
	ctx context.Context,
	o *options.Options,
	args []string,
) error {
	if err := p.IsSecurityOptionsAllowed(ctx); err != nil {
		return err
	}

	return p.parseInt(ctx, o, keys.MaxResultDimension, args...)
}

func (p *Parser) applyPresetOption(
	ctx context.Context,
	o *options.Options,
	args []string,
	usedPresets ...string,
) error {
	for _, preset := range args {
		if pr, ok := p.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				slog.Warn(fmt.Sprintf("Recursive preset usage is detected: %s", preset))
				continue
			}

			options.AppendToSlice(o, keys.UsedPresets, preset)

			if err := p.applyURLOptions(ctx, o, pr, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError(ctx, "preset", "Unknown preset: %s", preset)
		}
	}

	return nil
}
