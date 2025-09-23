package options

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

const maxClientHintDPR = 8

func (p *Parser) applyURLOption(po *Options, name string, args []string, usedPresets ...string) error {
	switch name {
	case "resize", "rs":
		return applyResizeOption(po, args)
	case "size", "s":
		return applySizeOption(po, args)
	case "resizing_type", "rt":
		return applyResizingTypeOption(po, args)
	case "width", "w":
		return applyWidthOption(po, args)
	case "height", "h":
		return applyHeightOption(po, args)
	case "min-width", "mw":
		return applyMinWidthOption(po, args)
	case "min-height", "mh":
		return applyMinHeightOption(po, args)
	case "zoom", "z":
		return applyZoomOption(po, args)
	case "dpr":
		return applyDprOption(po, args)
	case "enlarge", "el":
		return applyEnlargeOption(po, args)
	case "extend", "ex":
		return applyExtendOption(po, args)
	case "extend_aspect_ratio", "extend_ar", "exar":
		return applyExtendAspectRatioOption(po, args)
	case "gravity", "g":
		return applyGravityOption(po, args)
	case "crop", "c":
		return applyCropOption(po, args)
	case "trim", "t":
		return applyTrimOption(po, args)
	case "padding", "pd":
		return applyPaddingOption(po, args)
	case "auto_rotate", "ar":
		return applyAutoRotateOption(po, args)
	case "rotate", "rot":
		return applyRotateOption(po, args)
	case "background", "bg":
		return applyBackgroundOption(po, args)
	case "blur", "bl":
		return applyBlurOption(po, args)
	case "sharpen", "sh":
		return applySharpenOption(po, args)
	case "pixelate", "pix":
		return applyPixelateOption(po, args)
	case "watermark", "wm":
		return applyWatermarkOption(po, args)
	case "strip_metadata", "sm":
		return applyStripMetadataOption(po.Main(), args)
	case "keep_copyright", "kcr":
		return applyKeepCopyrightOption(po.Main(), args)
	case "strip_color_profile", "scp":
		return applyStripColorProfileOption(po.Main(), args)
	case "enforce_thumbnail", "eth":
		return applyEnforceThumbnailOption(po.Main(), args)
	// Saving options
	case "quality", "q":
		return applyQualityOption(po.Main(), args)
	case "format_quality", "fq":
		return applyFormatQualityOption(po.Main(), args)
	case "max_bytes", "mb":
		return applyMaxBytesOption(po.Main(), args)
	case "format", "f", "ext":
		return applyFormatOption(po.Main(), args)
	// Handling options
	case "skip_processing", "skp":
		return applySkipProcessingFormatsOption(po.Main(), args)
	case "raw":
		return applyRawOption(po.Main(), args)
	case "cachebuster", "cb":
		return applyCacheBusterOption(po.Main(), args)
	case "expires", "exp":
		return applyExpiresOption(po.Main(), args)
	case "filename", "fn":
		return applyFilenameOption(po.Main(), args)
	case "return_attachment", "att":
		return applyReturnAttachmentOption(po.Main(), args)
	// Presets
	case "preset", "pr":
		return applyPresetOption(p, po, args, usedPresets...)
	// Security
	case "max_src_resolution", "msr":
		return applyMaxSrcResolutionOption(p, po.Main(), args)
	case "max_src_file_size", "msfs":
		return applyMaxSrcFileSizeOption(p, po.Main(), args)
	case "max_animation_frames", "maf":
		return applyMaxAnimationFramesOption(p, po.Main(), args)
	case "max_animation_frame_resolution", "mafr":
		return applyMaxAnimationFrameResolutionOption(p, po.Main(), args)
	case "max_result_dimension", "mrd":
		return applyMaxResultDimensionOption(p, po.Main(), args)
	}

	return newUnknownOptionError("processing", name)
}

func (p *Parser) applyURLOptions(po *Options, options urlOptions, allowAll bool, usedPresets ...string) error {
	allowAll = allowAll || len(p.config.AllowedProcessingOptions) == 0

	for _, opt := range options {
		if !allowAll && !slices.Contains(p.config.AllowedProcessingOptions, opt.Name) {
			return newForbiddenOptionError("processing", opt.Name)
		}

		if err := p.applyURLOption(po, opt.Name, opt.Args, usedPresets...); err != nil {
			return err
		}
	}

	return nil
}

func (p *Parser) defaultProcessingOptions(headers http.Header) (*Options, error) {
	po := New()

	headerAccept := headers.Get("Accept")

	if (p.config.AutoWebp || p.config.EnforceWebp) && strings.Contains(headerAccept, "image/webp") {
		po.Set(keys.PreferWebP, true)

		if p.config.EnforceWebp {
			po.Set(keys.EnforceWebP, true)
		}
	}

	if (p.config.AutoAvif || p.config.EnforceAvif) && strings.Contains(headerAccept, "image/avif") {
		po.Set(keys.PreferAvif, true)

		if p.config.EnforceAvif {
			po.Set(keys.EnforceAvif, true)
		}
	}

	if (p.config.AutoJxl || p.config.EnforceJxl) && strings.Contains(headerAccept, "image/jxl") {
		po.Set(keys.PreferJxl, true)

		if p.config.EnforceJxl {
			po.Set(keys.EnforceJxl, true)
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
				po.Set(keys.Dpr, dpr)
			}
		}

		headerWidth := headers.Get("Sec-CH-Width")
		if len(headerWidth) == 0 {
			headerWidth = headers.Get("Width")
		}
		if len(headerWidth) > 0 {
			if w, err := strconv.Atoi(headerWidth); err == nil {
				po.Set(keys.Width, imath.Shrink(w, dpr))
			}
		}
	}

	if _, ok := p.presets["default"]; ok {
		if err := applyPresetOption(p, po, []string{"default"}); err != nil {
			return po, err
		}
	}

	return po, nil
}

// ParsePath parses the given request path and returns the processing options and image URL
func (p *Parser) ParsePath(
	path string,
	headers http.Header,
) (po *Options, imageURL string, err error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if p.config.OnlyPresets {
		po, imageURL, err = p.parsePathPresets(parts, headers)
	} else {
		po, imageURL, err = p.parsePathOptions(parts, headers)
	}

	if err != nil {
		return nil, "", ierrors.Wrap(err, 0)
	}

	return po, imageURL, nil
}

// parsePathOptions parses processing options from the URL path
func (p *Parser) parsePathOptions(parts []string, headers http.Header) (*Options, string, error) {
	if _, ok := resizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	po, err := p.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	options, urlParts := p.parseURLOptions(parts)

	if err = p.applyURLOptions(po, options, false); err != nil {
		return nil, "", err
	}

	url, extension, err := p.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !Get(po, keys.Raw, false) && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}

// parsePathPresets parses presets from the URL path
func (p *Parser) parsePathPresets(parts []string, headers http.Header) (*Options, string, error) {
	po, err := p.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], p.config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = applyPresetOption(p, po, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := p.DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !Get(po, keys.Raw, false) && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}
