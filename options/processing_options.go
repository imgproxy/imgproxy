package options

import (
	"encoding/base64"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/structdiff"
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

type FlipOptions struct {
	Horizontal bool
	Vertical   bool
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
	Page              int
	Gravity           GravityOptions
	Enlarge           bool
	Extend            ExtendOptions
	ExtendAspectRatio ExtendOptions
	Crop              CropOptions
	Padding           PaddingOptions
	Trim              TrimOptions
	Rotate            int
	Flip              FlipOptions
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
}

func NewProcessingOptions() *ProcessingOptions {
	po := ProcessingOptions{
		ResizingType:      ResizeFit,
		Width:             0,
		Height:            0,
		ZoomWidth:         1,
		ZoomHeight:        1,
		Gravity:           GravityOptions{Type: GravityCenter},
		Enlarge:           false,
		Extend:            ExtendOptions{Enabled: false, Gravity: GravityOptions{Type: GravityCenter}},
		ExtendAspectRatio: ExtendOptions{Enabled: false, Gravity: GravityOptions{Type: GravityCenter}},
		Padding:           PaddingOptions{Enabled: false},
		Trim:              TrimOptions{Enabled: false, Threshold: 10, Smart: true},
		Rotate:            0,
		Flip:              FlipOptions{Horizontal: false, Vertical: false},
		Quality:           0,
		MaxBytes:          0,
		Format:            imagetype.Unknown,
		Background:        vips.Color{R: 255, G: 255, B: 255},
		Blur:              0,
		Sharpen:           0,
		Dpr:               1,
		Watermark:         WatermarkOptions{Opacity: 1, Position: GravityOptions{Type: GravityCenter}},
		StripMetadata:     config.StripMetadata,
		KeepCopyright:     config.KeepCopyright,
		StripColorProfile: config.StripColorProfile,
		AutoRotate:        config.AutoRotate,
		EnforceThumbnail:  config.EnforceThumbnail,
		ReturnAttachment:  config.ReturnAttachment,

		SkipProcessingFormats: append([]imagetype.Type(nil), config.SkipProcessingFormats...),
		UsedPresets:           make([]string, 0, len(config.Presets)),

		SecurityOptions: security.DefaultOptions(),
		Page:            -1,

		// Basically, we need this to update ETag when `IMGPROXY_QUALITY` is changed
		defaultQuality: config.Quality,
	}

	po.FormatQuality = make(map[imagetype.Type]int, len(config.FormatQuality))
	for k, v := range config.FormatQuality {
		po.FormatQuality[k] = v
	}

	return &po
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

func (po *ProcessingOptions) Diff() structdiff.Entries {
	return structdiff.Diff(NewProcessingOptions(), po)
}

func (po *ProcessingOptions) String() string {
	return po.Diff().String()
}

func (po *ProcessingOptions) MarshalJSON() ([]byte, error) {
	return po.Diff().MarshalJSON()
}

func parseDimension(d *int, name, arg string) error {
	if v, err := strconv.Atoi(arg); err == nil && v >= 0 {
		*d = v
	} else {
		return newOptionArgumentError("Invalid %s: %s", name, arg)
	}

	return nil
}

func parseBoolOption(str string) bool {
	b, err := strconv.ParseBool(str)

	if err != nil {
		log.Warningf("`%s` is not a valid boolean value. Treated as false", str)
	}

	return b
}

func isGravityOffcetValid(gravity GravityType, offset float64) bool {
	return gravity != GravityFocusPoint || (offset >= 0 && offset <= 1)
}

func parseGravity(g *GravityOptions, name string, args []string, allowedTypes []GravityType) error {
	nArgs := len(args)

	if t, ok := gravityTypes[args[0]]; ok && slices.Contains(allowedTypes, t) {
		g.Type = t
	} else {
		return newOptionArgumentError("Invalid %s: %s", name, args[0])
	}

	switch g.Type {
	case GravitySmart:
		if nArgs > 1 {
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}
		g.X, g.Y = 0.0, 0.0

	case GravityFocusPoint:
		if nArgs != 3 {
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}
		fallthrough

	default:
		if nArgs > 3 {
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}

		if nArgs > 1 {
			if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffcetValid(g.Type, x) {
				g.X = x
			} else {
				return newOptionArgumentError("Invalid %s X: %s", name, args[1])
			}
		}

		if nArgs > 2 {
			if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffcetValid(g.Type, y) {
				g.Y = y
			} else {
				return newOptionArgumentError("Invalid %s Y: %s", name, args[2])
			}
		}
	}

	return nil
}

func parseExtend(opts *ExtendOptions, name string, args []string) error {
	if len(args) > 4 {
		return newOptionArgumentError("Invalid %s arguments: %v", name, args)
	}

	opts.Enabled = parseBoolOption(args[0])

	if len(args) > 1 {
		return parseGravity(&opts.Gravity, name+" gravity", args[1:], extendGravityTypes)
	}

	return nil
}

func applyWidthOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid width arguments: %v", args)
	}

	return parseDimension(&po.Width, "width", args[0])
}

func applyHeightOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid height arguments: %v", args)
	}

	return parseDimension(&po.Height, "height", args[0])
}

func applyMinWidthOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid min width arguments: %v", args)
	}

	return parseDimension(&po.MinWidth, "min width", args[0])
}

func applyMinHeightOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid min height arguments: %v", args)
	}

	return parseDimension(&po.MinHeight, " min height", args[0])
}

func applyEnlargeOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid enlarge arguments: %v", args)
	}

	po.Enlarge = parseBoolOption(args[0])

	return nil
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

	if z, err := strconv.ParseFloat(args[0], 64); err == nil && z > 0 {
		po.ZoomWidth = z
		po.ZoomHeight = z
	} else {
		return newOptionArgumentError("Invalid zoom value: %s", args[0])
	}

	if nArgs > 1 {
		if z, err := strconv.ParseFloat(args[1], 64); err == nil && z > 0 {
			po.ZoomHeight = z
		} else {
			return newOptionArgumentError("Invalid zoom value: %s", args[1])
		}
	}

	return nil
}

func applyDprOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid dpr arguments: %v", args)
	}

	if d, err := strconv.ParseFloat(args[0], 64); err == nil && d > 0 {
		po.Dpr = d
	} else {
		return newOptionArgumentError("Invalid dpr: %s", args[0])
	}

	return nil
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
		po.Trim.EqualHor = parseBoolOption(args[2])
	}

	if nArgs > 3 && len(args[3]) > 0 {
		po.Trim.EqualVer = parseBoolOption(args[3])
	}

	return nil
}

func applyRotateOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid rotate arguments: %v", args)
	}

	if r, err := strconv.Atoi(args[0]); err == nil && r%90 == 0 {
		po.Rotate = r
	} else {
		return newOptionArgumentError("Invalid rotation angle: %s", args[0])
	}

	return nil
}

func applyFlipOption(po *ProcessingOptions, args []string) error {
	if len(args) > 2 {
		return newOptionArgumentError("Invalid flip arguments: %v", args)
	}

	if len(args[0]) > 0 {
		po.Flip.Horizontal = parseBoolOption(args[0])
	}
	if len(args) > 1 && len(args[1]) > 0 {
		po.Flip.Vertical = parseBoolOption(args[1])
	}

	return nil
}

func applyQualityOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid quality arguments: %v", args)
	}

	if q, err := strconv.Atoi(args[0]); err == nil && q >= 0 && q <= 100 {
		po.Quality = q
	} else {
		return newOptionArgumentError("Invalid quality: %s", args[0])
	}

	return nil
}

func applyFormatQualityOption(po *ProcessingOptions, args []string) error {
	argsLen := len(args)
	if len(args)%2 != 0 {
		return newOptionArgumentError("Missing quality for: %s", args[argsLen-1])
	}

	for i := 0; i < argsLen; i += 2 {
		f, ok := imagetype.Types[args[i]]
		if !ok {
			return newOptionArgumentError("Invalid image format: %s", args[i])
		}

		if q, err := strconv.Atoi(args[i+1]); err == nil && q >= 0 && q <= 100 {
			po.FormatQuality[f] = q
		} else {
			return newOptionArgumentError("Invalid quality for %s: %s", args[i], args[i+1])
		}
	}

	return nil
}

func applyMaxBytesOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid max_bytes arguments: %v", args)
	}

	if max, err := strconv.Atoi(args[0]); err == nil && max >= 0 {
		po.MaxBytes = max
	} else {
		return newOptionArgumentError("Invalid max_bytes: %s", args[0])
	}

	return nil
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
	if len(args) > 1 {
		return newOptionArgumentError("Invalid blur arguments: %v", args)
	}

	if b, err := strconv.ParseFloat(args[0], 32); err == nil && b >= 0 {
		po.Blur = float32(b)
	} else {
		return newOptionArgumentError("Invalid blur: %s", args[0])
	}

	return nil
}

func applySharpenOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid sharpen arguments: %v", args)
	}

	if s, err := strconv.ParseFloat(args[0], 32); err == nil && s >= 0 {
		po.Sharpen = float32(s)
	} else {
		return newOptionArgumentError("Invalid sharpen: %s", args[0])
	}

	return nil
}

func applyPixelateOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid pixelate arguments: %v", args)
	}

	if p, err := strconv.Atoi(args[0]); err == nil && p >= 0 {
		po.Pixelate = p
	} else {
		return newOptionArgumentError("Invalid pixelate: %s", args[0])
	}

	return nil
}

func applyPresetOption(po *ProcessingOptions, args []string, usedPresets ...string) error {
	for _, preset := range args {
		if p, ok := presets[preset]; ok {
			if slices.Contains(usedPresets, preset) {
				log.Warningf("Recursive preset usage is detected: %s", preset)
				continue
			}

			po.UsedPresets = append(po.UsedPresets, preset)

			if err := applyURLOptions(po, p, true, append(usedPresets, preset)...); err != nil {
				return err
			}
		} else {
			return newOptionArgumentError("Unknown preset: %s", preset)
		}
	}

	return nil
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

	if f, ok := imagetype.Types[args[0]]; ok {
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
		if f, ok := imagetype.Types[format]; ok {
			po.SkipProcessingFormats = append(po.SkipProcessingFormats, f)
		} else {
			return newOptionArgumentError("Invalid image format in skip processing: %s", format)
		}
	}

	return nil
}

func applyPageOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid page arguments: %v", args)
	}

	if p, err := strconv.Atoi(args[0]); err == nil && p >= 0 {
		po.Page = p
	} else {
		return newOptionArgumentError("Invalid page: %s", args[0])
	}

	return nil
}

func applyRawOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid return_attachment arguments: %v", args)
	}

	po.Raw = parseBoolOption(args[0])

	return nil
}

func applyFilenameOption(po *ProcessingOptions, args []string) error {
	if len(args) > 2 {
		return newOptionArgumentError("Invalid filename arguments: %v", args)
	}

	po.Filename = args[0]

	if len(args) > 1 && parseBoolOption(args[1]) {
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
	if len(args) > 1 {
		return newOptionArgumentError("Invalid strip metadata arguments: %v", args)
	}

	po.StripMetadata = parseBoolOption(args[0])

	return nil
}

func applyKeepCopyrightOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid keep copyright arguments: %v", args)
	}

	po.KeepCopyright = parseBoolOption(args[0])

	return nil
}

func applyStripColorProfileOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid strip color profile arguments: %v", args)
	}

	po.StripColorProfile = parseBoolOption(args[0])

	return nil
}

func applyAutoRotateOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid auto rotate arguments: %v", args)
	}

	po.AutoRotate = parseBoolOption(args[0])

	return nil
}

func applyEnforceThumbnailOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid enforce thumbnail arguments: %v", args)
	}

	po.EnforceThumbnail = parseBoolOption(args[0])

	return nil
}

func applyReturnAttachmentOption(po *ProcessingOptions, args []string) error {
	if len(args) > 1 {
		return newOptionArgumentError("Invalid return_attachment arguments: %v", args)
	}

	po.ReturnAttachment = parseBoolOption(args[0])

	return nil
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

	if x, err := strconv.Atoi(args[0]); err == nil {
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

	if x, err := strconv.Atoi(args[0]); err == nil && x > 0 {
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

	if x, err := strconv.ParseFloat(args[0], 64); err == nil {
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

	if x, err := strconv.Atoi(args[0]); err == nil {
		po.SecurityOptions.MaxResultDimension = x
	} else {
		return newOptionArgumentError("Invalid max_result_dimension: %s", args[0])
	}

	return nil
}

func applyURLOption(po *ProcessingOptions, name string, args []string, usedPresets ...string) error {
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
	case "flip", "fl":
		return applyFlipOption(po, args)
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
		return applyPresetOption(po, args, usedPresets...)
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
	case "page":
		return applyPageOption(po, args)
	}

	return newUnknownOptionError("processing", name)
}

func applyURLOptions(po *ProcessingOptions, options urlOptions, allowAll bool, usedPresets ...string) error {
	allowAll = allowAll || len(config.AllowedProcessiongOptions) == 0

	for _, opt := range options {
		if !allowAll && !slices.Contains(config.AllowedProcessiongOptions, opt.Name) {
			return newForbiddenOptionError("processing", opt.Name)
		}

		if err := applyURLOption(po, opt.Name, opt.Args, usedPresets...); err != nil {
			return err
		}
	}

	return nil
}

func defaultProcessingOptions(headers http.Header) (*ProcessingOptions, error) {
	po := NewProcessingOptions()

	headerAccept := headers.Get("Accept")

	if strings.Contains(headerAccept, "image/webp") {
		po.PreferWebP = config.AutoWebp || config.EnforceWebp
		po.EnforceWebP = config.EnforceWebp
	}

	if strings.Contains(headerAccept, "image/avif") {
		po.PreferAvif = config.AutoAvif || config.EnforceAvif
		po.EnforceAvif = config.EnforceAvif
	}

	if strings.Contains(headerAccept, "image/jxl") {
		po.PreferJxl = config.AutoJxl || config.EnforceJxl
		po.EnforceJxl = config.EnforceJxl
	}

	if config.EnableClientHints {
		headerDPR := headers.Get("Sec-CH-DPR")
		if len(headerDPR) == 0 {
			headerDPR = headers.Get("DPR")
		}
		if len(headerDPR) > 0 {
			if dpr, err := strconv.ParseFloat(headerDPR, 64); err == nil && (dpr > 0 && dpr <= maxClientHintDPR) {
				po.Dpr = dpr
			}
		}

		headerWidth := headers.Get("Sec-CH-Width")
		if len(headerWidth) == 0 {
			headerWidth = headers.Get("Width")
		}
		if len(headerWidth) > 0 {
			if w, err := strconv.Atoi(headerWidth); err == nil {
				po.Width = imath.Shrink(w, po.Dpr)
			}
		}
	}

	if _, ok := presets["default"]; ok {
		if err := applyPresetOption(po, []string{"default"}); err != nil {
			return po, err
		}
	}

	return po, nil
}

func parsePathOptions(parts []string, headers http.Header) (*ProcessingOptions, string, error) {
	if _, ok := resizeTypes[parts[0]]; ok {
		return nil, "", newInvalidURLError("It looks like you're using the deprecated basic URL format")
	}

	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	options, urlParts := parseURLOptions(parts)

	if err = applyURLOptions(po, options, false); err != nil {
		return nil, "", err
	}

	url, extension, err := DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !po.Raw && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}

func parsePathPresets(parts []string, headers http.Header) (*ProcessingOptions, string, error) {
	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return nil, "", err
	}

	presets := strings.Split(parts[0], config.ArgumentsSeparator)
	urlParts := parts[1:]

	if err = applyPresetOption(po, presets); err != nil {
		return nil, "", err
	}

	url, extension, err := DecodeURL(urlParts)
	if err != nil {
		return nil, "", err
	}

	if !po.Raw && len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return nil, "", err
		}
	}

	return po, url, nil
}

func ParsePath(path string, headers http.Header) (*ProcessingOptions, string, error) {
	if path == "" || path == "/" {
		return nil, "", newInvalidURLError("Invalid path: %s", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	var (
		imageURL string
		po       *ProcessingOptions
		err      error
	)

	if config.OnlyPresets {
		po, imageURL, err = parsePathPresets(parts, headers)
	} else {
		po, imageURL, err = parsePathOptions(parts, headers)
	}

	if err != nil {
		return nil, "", ierrors.Wrap(err, 0)
	}

	return po, imageURL, nil
}
