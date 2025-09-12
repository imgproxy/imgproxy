package options

import (
	"encoding/base64"
	"slices"
	"strconv"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/structdiff"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
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
	factory  *Factory
}

func (wo WatermarkOptions) ShouldReplicate() bool {
	return wo.Position.Type == GravityReplicate
}

func (wo *WatermarkOptions) NewDefaultProcessingOptions() *ProcessingOptions {
	if wo.factory == nil {
		return nil
	}

	return wo.factory.New()
}

type ProcessingOptions struct {
	ResizingType      ResizeType
	Width             int
	Height            int
	MinWidth          int
	MinHeight         int
	ZoomWidth         float64
	ZoomHeight        float64
	Dpr               float64
	Gravity           GravityOptions
	Enlarge           bool
	Extend            ExtendOptions
	ExtendAspectRatio ExtendOptions
	Crop              CropOptions
	Padding           PaddingOptions
	Trim              TrimOptions
	Rotate            int
	Format            imagetype.Type
	Quality           int
	FormatQuality     map[imagetype.Type]int
	MaxBytes          int
	Flatten           bool
	Background        vips.Color
	Blur              float32
	Sharpen           float32
	Pixelate          int
	StripMetadata     bool
	KeepCopyright     bool
	StripColorProfile bool
	AutoRotate        bool
	EnforceThumbnail  bool

	SkipProcessingFormats []imagetype.Type

	CacheBuster string

	Expires *time.Time

	Watermark WatermarkOptions

	PreferWebP  bool
	EnforceWebP bool
	PreferAvif  bool
	EnforceAvif bool
	PreferJxl   bool
	EnforceJxl  bool

	Filename         string
	ReturnAttachment bool

	Raw bool

	UsedPresets []string

	SecurityOptions security.Options

	defaultQuality int

	factory *Factory
}

func (po *ProcessingOptions) GetQuality() int {
	q := po.Quality

	if q == 0 {
		q = po.FormatQuality[po.Format]
	}

	if q == 0 {
		q = po.defaultQuality
	}

	return q
}

func (po *ProcessingOptions) NewDefault() *ProcessingOptions {
	if po.factory == nil {
		return nil
	}

	return po.factory.New()
}

func (po *ProcessingOptions) Diff() structdiff.Entries {
	return structdiff.Diff(po.NewDefault(), po)
}

func (po *ProcessingOptions) String() string {
	return po.Diff().String()
}

func (po *ProcessingOptions) MarshalJSON() ([]byte, error) {
	return po.Diff().MarshalJSON()
}

// applyNumberOption applies a single argument option which is a number
func applyNumberOption[T number](args []string, name string, value *T, validate ...func(T) bool) error {
	if len(args) > 1 {
		return newOptionArgumentError("invalid %s arguments: %v", name, args)
	}

	val, err := parseNumber[T](args[0])
	if err != nil {
		return newOptionArgumentError("invalid %s: %s", name, args[0])
	}

	if len(validate) > 0 {
		for _, v := range validate {
			if v != nil && !v(val) {
				return newOptionArgumentError("invalid %s: %s", name, args[0])
			}
		}
	}

	if value == nil {
		return newOptionArgumentError("invalid %s: nil pointer", name)
	}

	*value = val
	return nil
}

// applyBoolOption applies a boolean option using the provided setter function
func applyBoolOption(args []string, name string, value *bool) error {
	if len(args) > 1 {
		return newOptionArgumentError("invalid %s arguments: %v", name, args)
	}

	if value == nil {
		return newOptionArgumentError("invalid %s: nil pointer", name)
	}

	*value = parseBool(args[0])
	return nil
}

func applyWidthOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "width", &po.Width, func(value int) bool {
		return value >= 0
	})
}

func applyHeightOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "height", &po.Height, func(value int) bool {
		return value >= 0
	})
}

func applyMinWidthOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "min width", &po.MinWidth, func(value int) bool {
		return value >= 0
	})
}

func applyMinHeightOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "min height", &po.MinHeight, func(value int) bool {
		return value >= 0
	})
}

func applyEnlargeOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "enlarge", &po.Enlarge)
}

func applyExtendOption(po *ProcessingOptions, args []string) error {
	return parseExtend(&po.Extend, "extend", args)
}

func applyExtendAspectRatioOption(po *ProcessingOptions, args []string) error {
	return parseExtend(&po.ExtendAspectRatio, "extend_aspect_ratio", args)
}

func applySizeOption(po *ProcessingOptions, args []string) (err error) {
	if len(args) > 7 {
		return newOptionArgumentError("Invalid size arguments: %v", args)
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
	if len(args) > 1 {
		return newOptionArgumentError("Invalid resizing type arguments: %v", args)
	}

	if r, ok := resizeTypes[args[0]]; ok {
		po.ResizingType = r
	} else {
		return newOptionArgumentError("Invalid resize type: %s", args[0])
	}

	return nil
}

func applyResizeOption(po *ProcessingOptions, args []string) error {
	if len(args) > 8 {
		return newOptionArgumentError("Invalid resize arguments: %v", args)
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

	if nArgs > 2 {
		return newOptionArgumentError("Invalid zoom arguments: %v", args)
	}

	if z, err := parseNumber[float64](args[0]); err == nil && z > 0 {
		po.ZoomWidth = z
		po.ZoomHeight = z
	} else {
		return newOptionArgumentError("Invalid zoom value: %s", args[0])
	}

	if nArgs > 1 {
		if z, err := parseNumber[float64](args[1]); err == nil && z > 0 {
			po.ZoomHeight = z
		} else {
			return newOptionArgumentError("Invalid zoom value: %s", args[1])
		}
	}

	return nil
}

func applyDprOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "dpr", &po.Dpr, func(value float64) bool {
		return value > 0
	})
}

func applyGravityOption(po *ProcessingOptions, args []string) error {
	return parseGravity(&po.Gravity, "gravity", args, cropGravityTypes)
}

func applyCropOption(po *ProcessingOptions, args []string) error {
	if w, err := strconv.ParseFloat(args[0], 64); err == nil && w >= 0 {
		po.Crop.Width = w
	} else {
		return newOptionArgumentError("Invalid crop width: %s", args[0])
	}

	if len(args) > 1 {
		if h, err := strconv.ParseFloat(args[1], 64); err == nil && h >= 0 {
			po.Crop.Height = h
		} else {
			return newOptionArgumentError("Invalid crop height: %s", args[1])
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
		if err := parseDimension(&po.Padding.Top, "padding top (+all)", args[0]); err != nil {
			return err
		}
		po.Padding.Right = po.Padding.Top
		po.Padding.Bottom = po.Padding.Top
		po.Padding.Left = po.Padding.Top
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if err := parseDimension(&po.Padding.Right, "padding right (+left)", args[1]); err != nil {
			return err
		}
		po.Padding.Left = po.Padding.Right
	}

	if nArgs > 2 && len(args[2]) > 0 {
		if err := parseDimension(&po.Padding.Bottom, "padding bottom", args[2]); err != nil {
			return err
		}
	}

	if nArgs > 3 && len(args[3]) > 0 {
		if err := parseDimension(&po.Padding.Left, "padding left", args[3]); err != nil {
			return err
		}
	}

	if po.Padding.Top == 0 && po.Padding.Right == 0 && po.Padding.Bottom == 0 && po.Padding.Left == 0 {
		po.Padding.Enabled = false
	}

	return nil
}

func applyTrimOption(po *ProcessingOptions, args []string) error {
	nArgs := len(args)

	if nArgs > 4 {
		return newOptionArgumentError("Invalid trim arguments: %v", args)
	}

	if t, err := strconv.ParseFloat(args[0], 64); err == nil && t >= 0 {
		po.Trim.Enabled = true
		po.Trim.Threshold = t
	} else {
		return newOptionArgumentError("Invalid trim threshold: %s", args[0])
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if c, err := vips.ColorFromHex(args[1]); err == nil {
			po.Trim.Color = c
			po.Trim.Smart = false
		} else {
			return newOptionArgumentError("Invalid trim color: %s", args[1])
		}
	}

	if nArgs > 2 && len(args[2]) > 0 {
		po.Trim.EqualHor = parseBool(args[2])
	}

	if nArgs > 3 && len(args[3]) > 0 {
		po.Trim.EqualVer = parseBool(args[3])
	}

	return nil
}

func applyRotateOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid rotate arguments: %v", args)
	}

	if r, err := parseNumber[int](args[0]); err == nil && r%90 == 0 {
		po.Rotate = r
	} else {
		return newOptionArgumentError("Invalid rotation angle: %s", args[0])
	}

	return nil
}
func applyQualityOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "quality", &po.Quality, func(value int) bool {
		return value >= 0 && value <= 100
	})
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

		if q, err := parseNumber[int](args[i+1]); err == nil && q >= 0 && q <= 100 {
			po.FormatQuality[f] = q
		} else {
			return newOptionArgumentError("Invalid quality for %s: %s", args[i], args[i+1])
		}
	}

	return nil
}

func applyMaxBytesOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "max_bytes", &po.MaxBytes, func(value int) bool {
		return value >= 0
	})
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
	return applyNumberOption(args, "blur", &po.Blur, func(value float32) bool {
		return value >= 0
	})
}

func applySharpenOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "sharpen", &po.Sharpen, func(value float32) bool {
		return value >= 0
	})
}

func applyPixelateOption(po *ProcessingOptions, args []string) error {
	return applyNumberOption(args, "pixelate", &po.Pixelate, func(value int) bool {
		return value >= 0
	})
}

func applyWatermarkOption(po *ProcessingOptions, args []string) error {
	if len(args) > 7 {
		return newOptionArgumentError("Invalid watermark arguments: %v", args)
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
		if x, err := strconv.ParseFloat(args[2], 64); err == nil {
			po.Watermark.Position.X = x
		} else {
			return newOptionArgumentError("Invalid watermark X offset: %s", args[2])
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if y, err := strconv.ParseFloat(args[3], 64); err == nil {
			po.Watermark.Position.Y = y
		} else {
			return newOptionArgumentError("Invalid watermark Y offset: %s", args[3])
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if s, err := strconv.ParseFloat(args[4], 64); err == nil && s >= 0 {
			po.Watermark.Scale = s
		} else {
			return newOptionArgumentError("Invalid watermark scale: %s", args[4])
		}
	}

	return nil
}

func applyFormatOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid format arguments: %v", args)
	}

	if f, ok := imagetype.GetTypeByName(args[0]); ok {
		po.Format = f
	} else {
		return newOptionArgumentError("Invalid image format: %s", args[0])
	}

	return nil
}
func applyCacheBusterOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid cache buster arguments: %v", args)
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
	return applyBoolOption(args, "raw", &po.Raw)
}

func applyFilenameOption(po *ProcessingOptions, args []string) error {
	if len(args) > 2 {
		return newOptionArgumentError("Invalid filename arguments: %v", args)
	}

	po.Filename = args[0]

	if len(args) > 1 && parseBool(args[1]) {
		decoded, err := base64.RawURLEncoding.DecodeString(po.Filename)
		if err != nil {
			return newOptionArgumentError("Invalid filename encoding: %s", err)
		}

		po.Filename = string(decoded)
	}

	return nil
}

func applyExpiresOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid expires arguments: %v", args)
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
	return applyBoolOption(args, "strip metadata", &po.StripMetadata)
}

func applyKeepCopyrightOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "keep copyright", &po.KeepCopyright)
}

func applyStripColorProfileOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "strip color profile", &po.StripColorProfile)
}

func applyAutoRotateOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "auto rotate", &po.AutoRotate)
}

func applyEnforceThumbnailOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "enforce thumbnail", &po.EnforceThumbnail)
}

func applyReturnAttachmentOption(po *ProcessingOptions, args []string) error {
	return applyBoolOption(args, "return_attachment", &po.ReturnAttachment)
}

func applyMaxSrcResolutionOption(po *ProcessingOptions, args []string) error {
	if err := security.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_src_resolution arguments: %v", args)
	}

	if x, err := strconv.ParseFloat(args[0], 64); err == nil && x > 0 {
		po.SecurityOptions.MaxSrcResolution = int(x * 1000000)
	} else {
		return newOptionArgumentError("Invalid max_src_resolution: %s", args[0])
	}

	return nil
}

func applyMaxSrcFileSizeOption(po *ProcessingOptions, args []string) error {
	if err := security.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_src_file_size arguments: %v", args)
	}

	if x, err := parseNumber[int](args[0]); err == nil {
		po.SecurityOptions.MaxSrcFileSize = x
	} else {
		return newOptionArgumentError("Invalid max_src_file_size: %s", args[0])
	}

	return nil
}

func applyMaxAnimationFramesOption(po *ProcessingOptions, args []string) error {
	if err := security.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_animation_frames arguments: %v", args)
	}

	if x, err := parseNumber[int](args[0]); err == nil && x > 0 {
		po.SecurityOptions.MaxAnimationFrames = x
	} else {
		return newOptionArgumentError("Invalid max_animation_frames: %s", args[0])
	}

	return nil
}

func applyMaxAnimationFrameResolutionOption(po *ProcessingOptions, args []string) error {
	if err := security.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_animation_frame_resolution arguments: %v", args)
	}

	if x, err := parseNumber[float64](args[0]); err == nil {
		po.SecurityOptions.MaxAnimationFrameResolution = int(x * 1000000)
	} else {
		return newOptionArgumentError("Invalid max_animation_frame_resolution: %s", args[0])
	}

	return nil
}

func applyMaxResultDimensionOption(po *ProcessingOptions, args []string) error {
	if err := security.IsSecurityOptionsAllowed(); err != nil {
		return err
	}

	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_result_dimension arguments: %v", args)
	}

	if x, err := parseNumber[int](args[0]); err == nil {
		po.SecurityOptions.MaxResultDimension = x
	} else {
		return newOptionArgumentError("Invalid max_result_dimension: %s", args[0])
	}

	return nil
}

func (f *Factory) applyURLOption(po *ProcessingOptions, name string, args []string, usedPresets ...string) error {
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
		return f.applyPresetOption(po, args, usedPresets...)
	// Security
	case "max_src_resolution", "msr":
		return applyMaxSrcResolutionOption(po, args)
	case "max_src_file_size", "msfs":
		return applyMaxSrcFileSizeOption(po, args)
	case "max_animation_frames", "maf":
		return applyMaxAnimationFramesOption(po, args)
	case "max_animation_frame_resolution", "mafr":
		return applyMaxAnimationFrameResolutionOption(po, args)
	case "max_result_dimension", "mrd":
		return applyMaxResultDimensionOption(po, args)
	}

	return newUnknownOptionError("processing", name)
}

func (f *Factory) applyURLOptions(po *ProcessingOptions, options urlOptions, allowAll bool, usedPresets ...string) error {
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

func (f *Factory) applyPresetOption(po *ProcessingOptions, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if p, ok := f.presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				log.Warningf("Recursive preset usage is detected: %s", preset)
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
