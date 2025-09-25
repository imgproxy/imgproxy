package optionsparser

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
)

const maxClientHintDPR = 8

func (p *Parser) applyURLOption(
	o *options.Options,
	name string,
	args []string,
	usedPresets ...string,
) error {
	switch name {
	case "resize", "rs":
		return applyResizeOption(o, args)
	case "size", "s":
		return applySizeOption(o, args)
	case "resizing_type", "rt":
		return applyResizingTypeOption(o, args)
	case "width", "w":
		return applyWidthOption(o, args)
	case "height", "h":
		return applyHeightOption(o, args)
	case "min-width", "mw":
		return applyMinWidthOption(o, args)
	case "min-height", "mh":
		return applyMinHeightOption(o, args)
	case "zoom", "z":
		return applyZoomOption(o, args)
	case "dpr":
		return applyDprOption(o, args)
	case "enlarge", "el":
		return applyEnlargeOption(o, args)
	case "extend", "ex":
		return applyExtendOption(o, args)
	case "extend_aspect_ratio", "extend_ar", "exar":
		return applyExtendAspectRatioOption(o, args)
	case "gravity", "g":
		return applyGravityOption(o, args)
	case "crop", "c":
		return applyCropOption(o, args)
	case "trim", "t":
		return applyTrimOption(o, args)
	case "padding", "pd":
		return applyPaddingOption(o, args)
	case "auto_rotate", "ar":
		return applyAutoRotateOption(o, args)
	case "rotate", "rot":
		return applyRotateOption(o, args)
	case "background", "bg":
		return applyBackgroundOption(o, args)
	case "blur", "bl":
		return applyBlurOption(o, args)
	case "sharpen", "sh":
		return applySharpenOption(o, args)
	case "pixelate", "pix":
		return applyPixelateOption(o, args)
	case "watermark", "wm":
		return applyWatermarkOption(o, args)
	case "strip_metadata", "sm":
		return applyStripMetadataOption(o.Main(), args)
	case "keep_copyright", "kcr":
		return applyKeepCopyrightOption(o.Main(), args)
	case "strip_color_profile", "scp":
		return applyStripColorProfileOption(o.Main(), args)
	case "enforce_thumbnail", "eth":
		return applyEnforceThumbnailOption(o.Main(), args)
	// Saving options
	case "quality", "q":
		return applyQualityOption(o.Main(), args)
	case "format_quality", "fq":
		return applyFormatQualityOption(o.Main(), args)
	case "max_bytes", "mb":
		return applyMaxBytesOption(o.Main(), args)
	case "format", "f", "ext":
		return applyFormatOption(o.Main(), args)
	// Handling options
	case "skip_processing", "skp":
		return applySkipProcessingFormatsOption(o.Main(), args)
	case "raw":
		return applyRawOption(o.Main(), args)
	case "cachebuster", "cb":
		return applyCacheBusterOption(o.Main(), args)
	case "expires", "exp":
		return applyExpiresOption(o.Main(), args)
	case "filename", "fn":
		return applyFilenameOption(o.Main(), args)
	case "return_attachment", "att":
		return applyReturnAttachmentOption(o.Main(), args)
	// Presets
	case "preset", "pr":
		return applyPresetOption(p, o, args, usedPresets...)
	// Security
	case "max_src_resolution", "msr":
		return applyMaxSrcResolutionOption(p, o.Main(), args)
	case "max_src_file_size", "msfs":
		return applyMaxSrcFileSizeOption(p, o.Main(), args)
	case "max_animation_frames", "maf":
		return applyMaxAnimationFramesOption(p, o.Main(), args)
	case "max_animation_frame_resolution", "mafr":
		return applyMaxAnimationFrameResolutionOption(p, o.Main(), args)
	case "max_result_dimension", "mrd":
		return applyMaxResultDimensionOption(p, o.Main(), args)
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

func (p *Parser) defaultProcessingOptions(headers http.Header) (*options.Options, error) {
	o := options.New()

	headerAccept := headers.Get("Accept")

	if (p.config.AutoWebp || p.config.EnforceWebp) && strings.Contains(headerAccept, "image/webp") {
		o.Set(keys.PreferWebP, true)

		if p.config.EnforceWebp {
			o.Set(keys.EnforceWebP, true)
		}
	}

	if (p.config.AutoAvif || p.config.EnforceAvif) && strings.Contains(headerAccept, "image/avif") {
		o.Set(keys.PreferAvif, true)

		if p.config.EnforceAvif {
			o.Set(keys.EnforceAvif, true)
		}
	}

	if (p.config.AutoJxl || p.config.EnforceJxl) && strings.Contains(headerAccept, "image/jxl") {
		o.Set(keys.PreferJxl, true)

		if p.config.EnforceJxl {
			o.Set(keys.EnforceJxl, true)
		}
	}

	if p.config.EnableClientHints {
		dpr := 1.0

		headerDPR := headers.Get("Sec-CH-DPR")
		if len(headerDPR) == 0 {
			headerDPR = headers.Get("DPR")
		}
		if len(headerDPR) > 0 {
			if d, err := strconv.ParseFloat(headerDPR, 64); err == nil && (d > 0 && d <= maxClientHintDPR) {
				dpr = d
				o.Set(keys.Dpr, dpr)
			}
		}

		headerWidth := headers.Get("Sec-CH-Width")
		if len(headerWidth) == 0 {
			headerWidth = headers.Get("Width")
		}
		if len(headerWidth) > 0 {
			if w, err := strconv.Atoi(headerWidth); err == nil {
				o.Set(keys.Width, imath.Shrink(w, dpr))
			}
		}
	}

	if _, ok := p.presets["default"]; ok {
		if err := applyPresetOption(p, o, []string{"default"}); err != nil {
			return o, err
		}
	}

	return o, nil
}

// ParsePath parses the given request path and returns the processing options and image URL
func (p *Parser) ParsePath(
	path string,
	headers http.Header,
) (o *options.Options, imageURL string, err error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if p.config.OnlyPresets {
		o, imageURL, err = p.parsePathPresets(parts, headers)
	} else {
		o, imageURL, err = p.parsePathOptions(parts, headers)
	}

	if err != nil {
		return nil, "", ierrors.Wrap(err, 0)
	}

	return o, imageURL, nil
}

// parsePathOptions parses processing options from the URL path
func (p *Parser) parsePathOptions(
	parts []string,
	headers http.Header,
) (*options.Options, string, error) {
	if _, ok := processing.ResizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	o, err := p.defaultProcessingOptions(headers)
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
		if err = applyFormatOption(o, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return o, url, nil
}

// parsePathPresets parses presets from the URL path
func (p *Parser) parsePathPresets(parts []string, headers http.Header) (*options.Options, string, error) {
	o, err := p.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], p.config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = applyPresetOption(p, o, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := p.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !options.Get(o, keys.Raw, false) && len(extension) > 0 {
		if err = applyFormatOption(o, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return o, url, nil
}
