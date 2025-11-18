package optionsparser

import (
	"slices"
	"strings"

	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
)

func (p *Parser) applyURLOption(
	o *options.Options,
	name string,
	args []string,
	usedPresets ...string,
) error {
	switch name {
	case "resize", "rs":
		return p.applyResizeOption(o, args)
	case "size", "s":
		return p.applySizeOption(o, args)
	case "resizing_type", "rt":
		return p.applyResizingTypeOption(o, args)
	case "width", "w":
		return p.applyWidthOption(o, args)
	case "height", "h":
		return p.applyHeightOption(o, args)
	case "min-width", "mw":
		return p.applyMinWidthOption(o, args)
	case "min-height", "mh":
		return p.applyMinHeightOption(o, args)
	case "zoom", "z":
		return p.applyZoomOption(o, args)
	case "dpr":
		return p.applyDprOption(o, args)
	case "enlarge", "el":
		return p.applyEnlargeOption(o, args)
	case "extend", "ex":
		return p.applyExtendOption(o, args)
	case "extend_aspect_ratio", "extend_ar", "exar":
		return p.applyExtendAspectRatioOption(o, args)
	case "gravity", "g":
		return p.applyGravityOption(o, args)
	case "crop", "c":
		return p.applyCropOption(o, args)
	case "trim", "t":
		return p.applyTrimOption(o, args)
	case "padding", "pd":
		return p.applyPaddingOption(o, args)
	case "auto_rotate", "ar":
		return p.applyAutoRotateOption(o, args)
	case "rotate", "rot":
		return p.applyRotateOption(o, args)
	case "flip", "fl":
		return p.applyFlipOption(o, args)
	case "background", "bg":
		return p.applyBackgroundOption(o, args)
	case "blur", "bl":
		return p.applyBlurOption(o, args)
	case "sharpen", "sh":
		return p.applySharpenOption(o, args)
	case "pixelate", "pix":
		return p.applyPixelateOption(o, args)
	case "watermark", "wm":
		return p.applyWatermarkOption(o, args)
	case "strip_metadata", "sm":
		return p.applyStripMetadataOption(o.Main(), args)
	case "keep_copyright", "kcr":
		return p.applyKeepCopyrightOption(o.Main(), args)
	case "strip_color_profile", "scp":
		return p.applyStripColorProfileOption(o.Main(), args)
	case "enforce_thumbnail", "eth":
		return p.applyEnforceThumbnailOption(o.Main(), args)
	// Saving options
	case "quality", "q":
		return p.applyQualityOption(o.Main(), args)
	case "format_quality", "fq":
		return p.applyFormatQualityOption(o.Main(), args)
	case "max_bytes", "mb":
		return p.applyMaxBytesOption(o.Main(), args)
	case "format", "f", "ext":
		return p.applyFormatOption(o.Main(), args)
	// Handling options
	case "skip_processing", "skp":
		return p.applySkipProcessingFormatsOption(o.Main(), args)
	case "raw":
		return p.applyRawOption(o.Main(), args)
	case "cachebuster", "cb":
		return p.applyCacheBusterOption(o.Main(), args)
	case "expires", "exp":
		return p.applyExpiresOption(o.Main(), args)
	case "filename", "fn":
		return p.applyFilenameOption(o.Main(), args)
	case "return_attachment", "att":
		return p.applyReturnAttachmentOption(o.Main(), args)
	// Presets
	case "preset", "pr":
		return p.applyPresetOption(o, args, usedPresets...)
	// Security
	case "max_src_resolution", "msr":
		return p.applyMaxSrcResolutionOption(o, args)
	case "max_src_file_size", "msfs":
		return p.applyMaxSrcFileSizeOption(o, args)
	case "max_animation_frames", "maf":
		return p.applyMaxAnimationFramesOption(o.Main(), args)
	case "max_animation_frame_resolution", "mafr":
		return p.applyMaxAnimationFrameResolutionOption(o.Main(), args)
	case "max_result_dimension", "mrd":
		return p.applyMaxResultDimensionOption(o.Main(), args)
	}

	return newUnknownOptionError("processing", name)
}

func (p *Parser) applyURLOptions(
	o *options.Options,
	options urlOptions,
	allowAll bool,
	usedPresets ...string,
) error {
	allowAll = allowAll || len(p.config.AllowedProcessingOptions) == 0

	for _, opt := range options {
		if !allowAll && !slices.Contains(p.config.AllowedProcessingOptions, opt.Name) {
			return newForbiddenOptionError("processing", opt.Name)
		}

		if err := p.applyURLOption(o, opt.Name, opt.Args, usedPresets...); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) defaultProcessingOptions(
	features *clientfeatures.Features,
) (*options.Options, error) {
	o := options.New()

	if features != nil {
		if features.PreferWebP || features.EnforceWebP {
			o.Set(keys.PreferWebP, true)
		}

		if features.EnforceWebP {
			o.Set(keys.EnforceWebP, true)
		}

		if features.PreferAvif || features.EnforceAvif {
			o.Set(keys.PreferAvif, true)
		}

		if features.EnforceAvif {
			o.Set(keys.EnforceAvif, true)
		}

		if features.PreferJxl || features.EnforceJxl {
			o.Set(keys.PreferJxl, true)
		}

		if features.EnforceJxl {
			o.Set(keys.EnforceJxl, true)
		}

		dpr := 1.0

		if features.ClientHintsDPR > 0 {
			o.Set(keys.Dpr, features.ClientHintsDPR)
			dpr = features.ClientHintsDPR
		}

		if features.ClientHintsWidth > 0 {
			o.Set(keys.Width, imath.Shrink(features.ClientHintsWidth, dpr))
		}
	}

	if _, ok := p.presets["default"]; ok {
		if err := p.applyPresetOption(o, []string{"default"}); err != nil {
			return o, err
		}
	}

	return o, nil
}

// ParsePath parses the given request path and returns the processing options and image URL
func (p *Parser) ParsePath(
	path string,
	features *clientfeatures.Features,
) (o *options.Options, imageURL string, err error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if p.config.OnlyPresets {
		o, imageURL, err = p.parsePathPresets(parts, features)
	} else {
		o, imageURL, err = p.parsePathOptions(parts, features)
	}

	if err != nil {
		return nil, "", errctx.Wrap(err, 0)
	}

	return o, imageURL, nil
}

// parsePathOptions parses processing options from the URL path
func (p *Parser) parsePathOptions(
	parts []string,
	features *clientfeatures.Features,
) (*options.Options, string, error) {
	if _, ok := processing.ResizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	o, err := p.defaultProcessingOptions(features)
	if err != nil {
		return nil, "", err
	}

	urlOpts, urlParts := p.parseURLOptions(parts)

	if err = p.applyURLOptions(o, urlOpts, false); err != nil {
		return nil, "", err
	}

	url, extension, err := p.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !options.Get(o, keys.Raw, false) && len(extension) > 0 {
		if err = p.applyFormatOption(o, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return o, url, nil
}

// parsePathPresets parses presets from the URL path
func (p *Parser) parsePathPresets(
	parts []string,
	features *clientfeatures.Features,
) (*options.Options, string, error) {
	o, err := p.defaultProcessingOptions(features)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], p.config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = p.applyPresetOption(o, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := p.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !options.Get(o, keys.Raw, false) && len(extension) > 0 {
		if err = p.applyFormatOption(o, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return o, url, nil
}
