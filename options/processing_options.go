package options

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
)

const maxClientHintDPR = 8

type ExtendOptions struct {
	Enabled bool
	Gravity GravityOptions
}

type CropOptions struct {
	Width   float64
	Height  float64
	Gravity GravityOptions
}

type PaddingOptions struct {
	Enabled bool
	Top     int
	Right   int
	Bottom  int
	Left    int
}

type TrimOptions struct {
	Enabled   bool
	Threshold float64
	Smart     bool
	Color     vips.Color
	EqualHor  bool
	EqualVer  bool
}

type WatermarkOptions struct {
	Enabled  bool
	Opacity  float64
	Position GravityOptions
	Scale    float64
}

func (wo WatermarkOptions) ShouldReplicate() bool {
	return wo.Position.Type == GravityReplicate
}

func (f *Factory) applyURLOption(po Options, name string, args []string, usedPresets ...string) error {
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
		return applyStripMetadataOption(po, args)
	case "keep_copyright", "kcr":
		return applyKeepCopyrightOption(po, args)
	case "strip_color_profile", "scp":
		return applyStripColorProfileOption(po, args)
	case "enforce_thumbnail", "eth":
		return applyEnforceThumbnailOption(po, args)
	// Saving options
	case "quality", "q":
		return applyQualityOption(po, args)
	case "format_quality", "fq":
		return applyFormatQualityOption(po, args)
	case "max_bytes", "mb":
		return applyMaxBytesOption(po, args)
	case "format", "f", "ext":
		return applyFormatOption(po, args)
	// Handling options
	case "skip_processing", "skp":
		return applySkipProcessingFormatsOption(po, args)
	case "raw":
		return applyRawOption(po, args)
	case "cachebuster", "cb":
		return applyCacheBusterOption(po, args)
	case "expires", "exp":
		return applyExpiresOption(po, args)
	case "filename", "fn":
		return applyFilenameOption(po, args)
	case "return_attachment", "att":
		return applyReturnAttachmentOption(po, args)
	// Presets
	case "preset", "pr":
		return applyPresetOption(f, po, args, usedPresets...)
	// Security
	case "max_src_resolution", "msr":
		return applyMaxSrcResolutionOption(f, po, args)
	case "max_src_file_size", "msfs":
		return applyMaxSrcFileSizeOption(f, po, args)
	case "max_animation_frames", "maf":
		return applyMaxAnimationFramesOption(f, po, args)
	case "max_animation_frame_resolution", "mafr":
		return applyMaxAnimationFrameResolutionOption(f, po, args)
	case "max_result_dimension", "mrd":
		return applyMaxResultDimensionOption(f, po, args)
	}

	return newUnknownOptionError("processing", name)
}

func (f *Factory) applyURLOptions(po Options, options urlOptions, allowAll bool, usedPresets ...string) error {
	allowAll = allowAll || len(f.config.AllowedProcessingOptions) == 0

	for _, opt := range options {
		if !allowAll && !slices.Contains(f.config.AllowedProcessingOptions, opt.Name) {
			return newForbiddenOptionError("processing", opt.Name)
		}

		if err := f.applyURLOption(po, opt.Name, opt.Args, usedPresets...); err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) defaultProcessingOptions(headers http.Header) (Options, error) {
	po := New()

	headerAccept := headers.Get("Accept")

	if (f.config.AutoWebp || f.config.EnforceWebp) && strings.Contains(headerAccept, "image/webp") {
		po[keys.PreferWebP] = true

		if f.config.EnforceWebp {
			po[keys.EnforceWebP] = true
		}
	}

	if (f.config.AutoAvif || f.config.EnforceAvif) && strings.Contains(headerAccept, "image/avif") {
		po[keys.PreferAvif] = true

		if f.config.EnforceAvif {
			po[keys.EnforceAvif] = true
		}
	}

	if (f.config.AutoJxl || f.config.EnforceJxl) && strings.Contains(headerAccept, "image/jxl") {
		po[keys.PreferJxl] = true

		if f.config.EnforceJxl {
			po[keys.EnforceJxl] = true
		}
	}

	if f.config.EnableClientHints {
		dpr := 1.0

		headerDPR := headers.Get("Sec-CH-DPR")
		if len(headerDPR) == 0 {
			headerDPR = headers.Get("DPR")
		}
		if len(headerDPR) > 0 {
			if d, err := strconv.ParseFloat(headerDPR, 64); err == nil && (d > 0 && d <= maxClientHintDPR) {
				dpr = d
				po[keys.Dpr] = dpr
			}
		}

		headerWidth := headers.Get("Sec-CH-Width")
		if len(headerWidth) == 0 {
			headerWidth = headers.Get("Width")
		}
		if len(headerWidth) > 0 {
			if w, err := strconv.Atoi(headerWidth); err == nil {
				po[keys.Width] = imath.Shrink(w, dpr)
			}
		}
	}

	if _, ok := f.presets["default"]; ok {
		if err := applyPresetOption(f, po, []string{"default"}); err != nil {
			return po, err
		}
	}

	return po, nil
}

// ParsePath parses the given request path and returns the processing options and image URL
func (f *Factory) ParsePath(
	path string,
	headers http.Header,
) (po Options, imageURL string, err error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if f.config.OnlyPresets {
		po, imageURL, err = f.parsePathPresets(parts, headers)
	} else {
		po, imageURL, err = f.parsePathOptions(parts, headers)
	}

	if err != nil {
		return nil, "", ierrors.Wrap(err, 0)
	}

	return po, imageURL, nil
}

// parsePathOptions parses processing options from the URL path
func (f *Factory) parsePathOptions(parts []string, headers http.Header) (Options, string, error) {
	if _, ok := resizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	po, err := f.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	options, urlParts := f.parseURLOptions(parts)

	if err = f.applyURLOptions(po, options, false); err != nil {
		return nil, "", err
	}

	url, extension, err := f.DecodeURL(urlParts)
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
func (f *Factory) parsePathPresets(parts []string, headers http.Header) (Options, string, error) {
	po, err := f.defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], f.config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = applyPresetOption(f, po, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := f.DecodeURL(urlParts)
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
