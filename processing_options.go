package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v2/structdiff"
)

type urlOption struct {
	Name string
	Args []string
}
type urlOptions []urlOption

type processingHeaders struct {
	Accept        string
	Width         string
	ViewportWidth string
	DPR           string
}

type gravityType int

const (
	gravityUnknown gravityType = iota
	gravityCenter
	gravityNorth
	gravityEast
	gravitySouth
	gravityWest
	gravityNorthWest
	gravityNorthEast
	gravitySouthWest
	gravitySouthEast
	gravitySmart
	gravityFocusPoint
)

var gravityTypes = map[string]gravityType{
	"ce":   gravityCenter,
	"no":   gravityNorth,
	"ea":   gravityEast,
	"so":   gravitySouth,
	"we":   gravityWest,
	"nowe": gravityNorthWest,
	"noea": gravityNorthEast,
	"sowe": gravitySouthWest,
	"soea": gravitySouthEast,
	"sm":   gravitySmart,
	"fp":   gravityFocusPoint,
}

type resizeType int

const (
	resizeFit resizeType = iota
	resizeFill
	resizeCrop
	resizeAuto
)

var resizeTypes = map[string]resizeType{
	"fit":  resizeFit,
	"fill": resizeFill,
	"crop": resizeCrop,
	"auto": resizeAuto,
}

type rgbColor struct{ R, G, B uint8 }

var hexColorRegex = regexp.MustCompile("^([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$")

const (
	hexColorLongFormat  = "%02x%02x%02x"
	hexColorShortFormat = "%1x%1x%1x"
)

type gravityOptions struct {
	Type gravityType
	X, Y float64
}

type extendOptions struct {
	Enabled bool
	Gravity gravityOptions
}

type cropOptions struct {
	Width   float64
	Height  float64
	Gravity gravityOptions
}

type paddingOptions struct {
	Enabled bool
	Top     int
	Right   int
	Bottom  int
	Left    int
}

type trimOptions struct {
	Enabled   bool
	Threshold float64
	Smart     bool
	Color     rgbColor
	EqualHor  bool
	EqualVer  bool
}

type watermarkOptions struct {
	Enabled   bool
	Opacity   float64
	Replicate bool
	Gravity   gravityOptions
	Scale     float64
}

type processingOptions struct {
	ResizingType      resizeType
	Width             int
	Height            int
	Dpr               float64
	Gravity           gravityOptions
	Enlarge           bool
	Extend            extendOptions
	Crop              cropOptions
	Padding           paddingOptions
	Trim              trimOptions
	Rotate            int
	Format            imageType
	Quality           int
	MaxBytes          int
	Flatten           bool
	Background        rgbColor
	Blur              float32
	Sharpen           float32
	StripMetadata     bool
	StripColorProfile bool
	AutoRotate        bool

	CacheBuster string

	Watermark watermarkOptions

	PreferWebP  bool
	EnforceWebP bool
	PreferAvif  bool
	EnforceAvif bool

	Filename string

	UsedPresets []string
}

const (
	imageURLCtxKey          = ctxKey("imageUrl")
	processingOptionsCtxKey = ctxKey("processingOptions")
	urlTokenPlain           = "plain"
	maxClientHintDPR        = 8

	msgForbidden     = "Forbidden"
	msgInvalidURL    = "Invalid URL"
	msgInvalidSource = "Invalid Source"
)

func (gt gravityType) String() string {
	for k, v := range gravityTypes {
		if v == gt {
			return k
		}
	}
	return ""
}

func (gt gravityType) MarshalJSON() ([]byte, error) {
	for k, v := range gravityTypes {
		if v == gt {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}

func (rt resizeType) String() string {
	for k, v := range resizeTypes {
		if v == rt {
			return k
		}
	}
	return ""
}

func (rt resizeType) MarshalJSON() ([]byte, error) {
	for k, v := range resizeTypes {
		if v == rt {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}

var (
	_newProcessingOptions    processingOptions
	newProcessingOptionsOnce sync.Once
)

func newProcessingOptions() *processingOptions {
	newProcessingOptionsOnce.Do(func() {
		_newProcessingOptions = processingOptions{
			ResizingType:      resizeFit,
			Width:             0,
			Height:            0,
			Gravity:           gravityOptions{Type: gravityCenter},
			Enlarge:           false,
			Extend:            extendOptions{Enabled: false, Gravity: gravityOptions{Type: gravityCenter}},
			Padding:           paddingOptions{Enabled: false},
			Trim:              trimOptions{Enabled: false, Threshold: 10, Smart: true},
			Rotate:            0,
			Quality:           0,
			MaxBytes:          0,
			Format:            imageTypeUnknown,
			Background:        rgbColor{255, 255, 255},
			Blur:              0,
			Sharpen:           0,
			Dpr:               1,
			Watermark:         watermarkOptions{Opacity: 1, Replicate: false, Gravity: gravityOptions{Type: gravityCenter}},
			StripMetadata:     conf.StripMetadata,
			StripColorProfile: conf.StripColorProfile,
			AutoRotate:        conf.AutoRotate,
		}
	})

	po := _newProcessingOptions
	po.UsedPresets = make([]string, 0, len(conf.Presets))

	return &po
}

func (po *processingOptions) getQuality() int {
	q := po.Quality

	if q == 0 {
		q = conf.FormatQuality[po.Format]
	}

	if q == 0 {
		q = conf.Quality
	}

	return q
}

func (po *processingOptions) isPresetUsed(name string) bool {
	for _, usedName := range po.UsedPresets {
		if usedName == name {
			return true
		}
	}
	return false
}

func (po *processingOptions) presetUsed(name string) {
	po.UsedPresets = append(po.UsedPresets, name)
}

func (po *processingOptions) Diff() structdiff.Entries {
	return structdiff.Diff(newProcessingOptions(), po)
}

func (po *processingOptions) String() string {
	return po.Diff().String()
}

func (po *processingOptions) MarshalJSON() ([]byte, error) {
	return po.Diff().MarshalJSON()
}

func colorFromHex(hexcolor string) (rgbColor, error) {
	c := rgbColor{}

	if !hexColorRegex.MatchString(hexcolor) {
		return c, fmt.Errorf("Invalid hex color: %s", hexcolor)
	}

	if len(hexcolor) == 3 {
		fmt.Sscanf(hexcolor, hexColorShortFormat, &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	} else {
		fmt.Sscanf(hexcolor, hexColorLongFormat, &c.R, &c.G, &c.B)
	}

	return c, nil
}

func addBaseURL(u string) string {
	if len(conf.BaseURL) == 0 || strings.HasPrefix(u, conf.BaseURL) {
		return u
	}

	return fmt.Sprintf("%s%s", conf.BaseURL, u)
}

func decodeBase64URL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "")
	urlParts := strings.Split(encoded, ".")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	imageURL, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return addBaseURL(string(imageURL)), format, nil
}

func decodePlainURL(parts []string) (string, string, error) {
	var format string

	encoded := strings.Join(parts, "/")
	urlParts := strings.Split(encoded, "@")

	if len(urlParts[0]) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if len(urlParts) > 2 {
		return "", "", fmt.Errorf("Multiple formats are specified: %s", encoded)
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	unescaped, err := url.PathUnescape(urlParts[0])
	if err != nil {
		return "", "", fmt.Errorf("Invalid url encoding: %s", encoded)
	}

	return addBaseURL(unescaped), format, nil
}

func decodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", errors.New("Image URL is empty")
	}

	if parts[0] == urlTokenPlain && len(parts) > 1 {
		return decodePlainURL(parts[1:])
	}

	return decodeBase64URL(parts)
}

func parseDimension(d *int, name, arg string) error {
	if v, err := strconv.Atoi(arg); err == nil && v >= 0 {
		*d = v
	} else {
		return fmt.Errorf("Invalid %s: %s", name, arg)
	}

	return nil
}

func parseBoolOption(str string) bool {
	b, err := strconv.ParseBool(str)

	if err != nil {
		logWarning("`%s` is not a valid boolean value. Treated as false", str)
	}

	return b
}

func isGravityOffcetValid(gravity gravityType, offset float64) bool {
	if gravity == gravityCenter {
		return true
	}

	return offset >= 0 && (gravity != gravityFocusPoint || offset <= 1)
}

func parseGravity(g *gravityOptions, args []string) error {
	nArgs := len(args)

	if nArgs > 3 {
		return fmt.Errorf("Invalid gravity arguments: %v", args)
	}

	if t, ok := gravityTypes[args[0]]; ok {
		g.Type = t
	} else {
		return fmt.Errorf("Invalid gravity: %s", args[0])
	}

	if g.Type == gravitySmart && nArgs > 1 {
		return fmt.Errorf("Invalid gravity arguments: %v", args)
	} else if g.Type == gravityFocusPoint && nArgs != 3 {
		return fmt.Errorf("Invalid gravity arguments: %v", args)
	}

	if nArgs > 1 {
		if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffcetValid(g.Type, x) {
			g.X = x
		} else {
			return fmt.Errorf("Invalid gravity X: %s", args[1])
		}
	}

	if nArgs > 2 {
		if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffcetValid(g.Type, y) {
			g.Y = y
		} else {
			return fmt.Errorf("Invalid gravity Y: %s", args[2])
		}
	}

	return nil
}

func applyWidthOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid width arguments: %v", args)
	}

	return parseDimension(&po.Width, "width", args[0])
}

func applyHeightOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid height arguments: %v", args)
	}

	return parseDimension(&po.Height, "height", args[0])
}

func applyEnlargeOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid enlarge arguments: %v", args)
	}

	po.Enlarge = parseBoolOption(args[0])

	return nil
}

func applyExtendOption(po *processingOptions, args []string) error {
	if len(args) > 4 {
		return fmt.Errorf("Invalid extend arguments: %v", args)
	}

	po.Extend.Enabled = parseBoolOption(args[0])

	if len(args) > 1 {
		if err := parseGravity(&po.Extend.Gravity, args[1:]); err != nil {
			return err
		}

		if po.Extend.Gravity.Type == gravitySmart {
			return errors.New("extend doesn't support smart gravity")
		}
	}

	return nil
}

func applySizeOption(po *processingOptions, args []string) (err error) {
	if len(args) > 7 {
		return fmt.Errorf("Invalid size arguments: %v", args)
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

func applyResizingTypeOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid resizing type arguments: %v", args)
	}

	if r, ok := resizeTypes[args[0]]; ok {
		po.ResizingType = r
	} else {
		return fmt.Errorf("Invalid resize type: %s", args[0])
	}

	return nil
}

func applyResizeOption(po *processingOptions, args []string) error {
	if len(args) > 8 {
		return fmt.Errorf("Invalid resize arguments: %v", args)
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

func applyDprOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid dpr arguments: %v", args)
	}

	if d, err := strconv.ParseFloat(args[0], 64); err == nil && d > 0 {
		po.Dpr = d
	} else {
		return fmt.Errorf("Invalid dpr: %s", args[0])
	}

	return nil
}

func applyGravityOption(po *processingOptions, args []string) error {
	return parseGravity(&po.Gravity, args)
}

func applyCropOption(po *processingOptions, args []string) error {
	if len(args) > 5 {
		return fmt.Errorf("Invalid crop arguments: %v", args)
	}

	if w, err := strconv.ParseFloat(args[0], 64); err == nil && w >= 0 {
		po.Crop.Width = w
	} else {
		return fmt.Errorf("Invalid crop width: %s", args[0])
	}

	if len(args) > 1 {
		if h, err := strconv.ParseFloat(args[1], 64); err == nil && h >= 0 {
			po.Crop.Height = h
		} else {
			return fmt.Errorf("Invalid crop height: %s", args[1])
		}
	}

	if len(args) > 2 {
		return parseGravity(&po.Crop.Gravity, args[2:])
	}

	return nil
}

func applyPaddingOption(po *processingOptions, args []string) error {
	nArgs := len(args)

	if nArgs < 1 || nArgs > 4 {
		return fmt.Errorf("Invalid padding arguments: %v", args)
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

func applyTrimOption(po *processingOptions, args []string) error {
	nArgs := len(args)

	if nArgs > 4 {
		return fmt.Errorf("Invalid trim arguments: %v", args)
	}

	if t, err := strconv.ParseFloat(args[0], 64); err == nil && t >= 0 {
		po.Trim.Enabled = true
		po.Trim.Threshold = t
	} else {
		return fmt.Errorf("Invalid trim threshold: %s", args[0])
	}

	if nArgs > 1 && len(args[1]) > 0 {
		if c, err := colorFromHex(args[1]); err == nil {
			po.Trim.Color = c
			po.Trim.Smart = false
		} else {
			return fmt.Errorf("Invalid trim color: %s", args[1])
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

func applyRotateOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid rotate arguments: %v", args)
	}

	if r, err := strconv.Atoi(args[0]); err == nil && r%90 == 0 {
		po.Rotate = r
	} else {
		return fmt.Errorf("Invalid rotation angle: %s", args[0])
	}

	return nil
}

func applyQualityOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid quality arguments: %v", args)
	}

	if q, err := strconv.Atoi(args[0]); err == nil && q >= 0 && q <= 100 {
		po.Quality = q
	} else {
		return fmt.Errorf("Invalid quality: %s", args[0])
	}

	return nil
}

func applyMaxBytesOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid max_bytes arguments: %v", args)
	}

	if max, err := strconv.Atoi(args[0]); err == nil && max >= 0 {
		po.MaxBytes = max
	} else {
		return fmt.Errorf("Invalid max_bytes: %s", args[0])
	}

	return nil
}

func applyBackgroundOption(po *processingOptions, args []string) error {
	switch len(args) {
	case 1:
		if len(args[0]) == 0 {
			po.Flatten = false
		} else if c, err := colorFromHex(args[0]); err == nil {
			po.Flatten = true
			po.Background = c
		} else {
			return fmt.Errorf("Invalid background argument: %s", err)
		}

	case 3:
		po.Flatten = true

		if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r <= 255 {
			po.Background.R = uint8(r)
		} else {
			return fmt.Errorf("Invalid background red channel: %s", args[0])
		}

		if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g <= 255 {
			po.Background.G = uint8(g)
		} else {
			return fmt.Errorf("Invalid background green channel: %s", args[1])
		}

		if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b <= 255 {
			po.Background.B = uint8(b)
		} else {
			return fmt.Errorf("Invalid background blue channel: %s", args[2])
		}

	default:
		return fmt.Errorf("Invalid background arguments: %v", args)
	}

	return nil
}

func applyBlurOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid blur arguments: %v", args)
	}

	if b, err := strconv.ParseFloat(args[0], 32); err == nil && b >= 0 {
		po.Blur = float32(b)
	} else {
		return fmt.Errorf("Invalid blur: %s", args[0])
	}

	return nil
}

func applySharpenOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid sharpen arguments: %v", args)
	}

	if s, err := strconv.ParseFloat(args[0], 32); err == nil && s >= 0 {
		po.Sharpen = float32(s)
	} else {
		return fmt.Errorf("Invalid sharpen: %s", args[0])
	}

	return nil
}

func applyPresetOption(po *processingOptions, args []string) error {
	for _, preset := range args {
		if p, ok := conf.Presets[preset]; ok {
			if po.isPresetUsed(preset) {
				logWarning("Recursive preset usage is detected: %s", preset)
				continue
			}

			po.presetUsed(preset)

			if err := applyProcessingOptions(po, p); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Unknown preset: %s", preset)
		}
	}

	return nil
}

func applyWatermarkOption(po *processingOptions, args []string) error {
	if len(args) > 7 {
		return fmt.Errorf("Invalid watermark arguments: %v", args)
	}

	if o, err := strconv.ParseFloat(args[0], 64); err == nil && o >= 0 && o <= 1 {
		po.Watermark.Enabled = o > 0
		po.Watermark.Opacity = o
	} else {
		return fmt.Errorf("Invalid watermark opacity: %s", args[0])
	}

	if len(args) > 1 && len(args[1]) > 0 {
		if args[1] == "re" {
			po.Watermark.Replicate = true
		} else if g, ok := gravityTypes[args[1]]; ok && g != gravityFocusPoint && g != gravitySmart {
			po.Watermark.Gravity.Type = g
		} else {
			return fmt.Errorf("Invalid watermark position: %s", args[1])
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if x, err := strconv.Atoi(args[2]); err == nil {
			po.Watermark.Gravity.X = float64(x)
		} else {
			return fmt.Errorf("Invalid watermark X offset: %s", args[2])
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if y, err := strconv.Atoi(args[3]); err == nil {
			po.Watermark.Gravity.Y = float64(y)
		} else {
			return fmt.Errorf("Invalid watermark Y offset: %s", args[3])
		}
	}

	if len(args) > 4 && len(args[4]) > 0 {
		if s, err := strconv.ParseFloat(args[4], 64); err == nil && s >= 0 {
			po.Watermark.Scale = s
		} else {
			return fmt.Errorf("Invalid watermark scale: %s", args[4])
		}
	}

	return nil
}

func applyFormatOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid format arguments: %v", args)
	}

	if f, ok := imageTypes[args[0]]; ok {
		po.Format = f
	} else {
		return fmt.Errorf("Invalid image format: %s", args[0])
	}

	if !imageTypeSaveSupport(po.Format) {
		return fmt.Errorf("Resulting image format is not supported: %s", po.Format)
	}

	return nil
}

func applyCacheBusterOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid cache buster arguments: %v", args)
	}

	po.CacheBuster = args[0]

	return nil
}

func applyFilenameOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid filename arguments: %v", args)
	}

	po.Filename = args[0]

	return nil
}

func applyStripMetadataOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid strip metadata arguments: %v", args)
	}

	po.StripMetadata = parseBoolOption(args[0])

	return nil
}

func applyStripColorProfileOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid strip color profile arguments: %v", args)
	}

	po.StripColorProfile = parseBoolOption(args[0])

	return nil
}

func applyAutoRotateOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid auto rotate arguments: %v", args)
	}

	po.AutoRotate = parseBoolOption(args[0])

	return nil
}

func applyProcessingOption(po *processingOptions, name string, args []string) error {
	switch name {
	case "format", "f", "ext":
		return applyFormatOption(po, args)
	case "resize", "rs":
		return applyResizeOption(po, args)
	case "resizing_type", "rt":
		return applyResizingTypeOption(po, args)
	case "size", "s":
		return applySizeOption(po, args)
	case "width", "w":
		return applyWidthOption(po, args)
	case "height", "h":
		return applyHeightOption(po, args)
	case "enlarge", "el":
		return applyEnlargeOption(po, args)
	case "extend", "ex":
		return applyExtendOption(po, args)
	case "dpr":
		return applyDprOption(po, args)
	case "gravity", "g":
		return applyGravityOption(po, args)
	case "crop", "c":
		return applyCropOption(po, args)
	case "trim", "t":
		return applyTrimOption(po, args)
	case "rotate", "rot":
		return applyRotateOption(po, args)
	case "padding", "pd":
		return applyPaddingOption(po, args)
	case "quality", "q":
		return applyQualityOption(po, args)
	case "max_bytes", "mb":
		return applyMaxBytesOption(po, args)
	case "background", "bg":
		return applyBackgroundOption(po, args)
	case "blur", "bl":
		return applyBlurOption(po, args)
	case "sharpen", "sh":
		return applySharpenOption(po, args)
	case "watermark", "wm":
		return applyWatermarkOption(po, args)
	case "preset", "pr":
		return applyPresetOption(po, args)
	case "cachebuster", "cb":
		return applyCacheBusterOption(po, args)
	case "strip_metadata", "sm":
		return applyStripMetadataOption(po, args)
	case "strip_color_profile", "scp":
		return applyStripColorProfileOption(po, args)
	case "auto_rotate", "ar":
		return applyAutoRotateOption(po, args)
	case "filename", "fn":
		return applyFilenameOption(po, args)
	}

	return fmt.Errorf("Unknown processing option: %s", name)
}

func applyProcessingOptions(po *processingOptions, options urlOptions) error {
	for _, opt := range options {
		if err := applyProcessingOption(po, opt.Name, opt.Args); err != nil {
			return err
		}
	}

	return nil
}

func isAllowedSource(imageURL string) bool {
	if len(conf.AllowedSources) == 0 {
		return true
	}
	for _, allowedSource := range conf.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return true
		}
	}
	return false
}

func parseURLOptions(opts []string) (urlOptions, []string) {
	parsed := make(urlOptions, 0, len(opts))
	urlStart := len(opts) + 1

	for i, opt := range opts {
		args := strings.Split(opt, ":")

		if len(args) == 1 {
			urlStart = i
			break
		}

		parsed = append(parsed, urlOption{Name: args[0], Args: args[1:]})
	}

	var rest []string

	if urlStart < len(opts) {
		rest = opts[urlStart:]
	} else {
		rest = []string{}
	}

	return parsed, rest
}

func defaultProcessingOptions(headers *processingHeaders) (*processingOptions, error) {
	po := newProcessingOptions()

	if strings.Contains(headers.Accept, "image/webp") {
		po.PreferWebP = conf.EnableWebpDetection || conf.EnforceWebp
		po.EnforceWebP = conf.EnforceWebp
	}

	if strings.Contains(headers.Accept, "image/avif") {
		po.PreferAvif = conf.EnableAvifDetection || conf.EnforceAvif
		po.EnforceAvif = conf.EnforceAvif
	}

	if conf.EnableClientHints && len(headers.ViewportWidth) > 0 {
		if vw, err := strconv.Atoi(headers.ViewportWidth); err == nil {
			po.Width = vw
		}
	}
	if conf.EnableClientHints && len(headers.Width) > 0 {
		if w, err := strconv.Atoi(headers.Width); err == nil {
			po.Width = w
		}
	}
	if conf.EnableClientHints && len(headers.DPR) > 0 {
		if dpr, err := strconv.ParseFloat(headers.DPR, 64); err == nil && (dpr > 0 && dpr <= maxClientHintDPR) {
			po.Dpr = dpr
		}
	}
	if _, ok := conf.Presets["default"]; ok {
		if err := applyPresetOption(po, []string{"default"}); err != nil {
			return po, err
		}
	}

	return po, nil
}

func parsePathAdvanced(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	options, urlParts := parseURLOptions(parts)

	if err = applyProcessingOptions(po, options); err != nil {
		return "", po, err
	}

	url, extension, err := decodeURL(urlParts)
	if err != nil {
		return "", po, err
	}

	if len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return "", po, err
		}
	}

	return url, po, nil
}

func parsePathPresets(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	presets := strings.Split(parts[0], ":")
	urlParts := parts[1:]

	if err = applyPresetOption(po, presets); err != nil {
		return "", nil, err
	}

	url, extension, err := decodeURL(urlParts)
	if err != nil {
		return "", po, err
	}

	if len(extension) > 0 {
		if err = applyFormatOption(po, []string{extension}); err != nil {
			return "", po, err
		}
	}

	return url, po, nil
}

func parsePathBasic(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	logWarning("The basic URL format is deprecated and can be removed in future versions. Use advanced URL format instead")

	if len(parts) < 6 {
		return "", nil, fmt.Errorf("Invalid basic URL format arguments: %s", strings.Join(parts, "/"))
	}

	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	po.ResizingType = resizeTypes[parts[0]]

	if err = applyWidthOption(po, parts[1:2]); err != nil {
		return "", po, err
	}

	if err = applyHeightOption(po, parts[2:3]); err != nil {
		return "", po, err
	}

	if err = applyGravityOption(po, strings.Split(parts[3], ":")); err != nil {
		return "", po, err
	}

	if err = applyEnlargeOption(po, parts[4:5]); err != nil {
		return "", po, err
	}

	url, extension, err := decodeURL(parts[5:])
	if err != nil {
		return "", po, err
	}

	if len(extension) > 0 {
		if err := applyFormatOption(po, []string{extension}); err != nil {
			return "", po, err
		}
	}

	return url, po, nil
}

func parsePath(ctx context.Context, r *http.Request) (context.Context, error) {
	var err error

	path := trimAfter(r.RequestURI, '?')

	if len(conf.PathPrefix) > 0 {
		path = strings.TrimPrefix(path, conf.PathPrefix)
	}

	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return ctx, newError(404, fmt.Sprintf("Invalid path: %s", path), msgInvalidURL)
	}

	if !conf.AllowInsecure {
		if err = validatePath(parts[0], strings.TrimPrefix(path, parts[0])); err != nil {
			return ctx, newError(403, err.Error(), msgForbidden)
		}
	}

	headers := &processingHeaders{
		Accept:        r.Header.Get("Accept"),
		Width:         r.Header.Get("Width"),
		ViewportWidth: r.Header.Get("Viewport-Width"),
		DPR:           r.Header.Get("DPR"),
	}

	var imageURL string
	var po *processingOptions

	if conf.OnlyPresets {
		imageURL, po, err = parsePathPresets(parts[1:], headers)
	} else if _, ok := resizeTypes[parts[1]]; ok {
		imageURL, po, err = parsePathBasic(parts[1:], headers)
	} else {
		imageURL, po, err = parsePathAdvanced(parts[1:], headers)
	}

	if err != nil {
		return ctx, newError(404, err.Error(), msgInvalidURL)
	}

	if !isAllowedSource(imageURL) {
		return ctx, newError(404, "Invalid source", msgInvalidSource)
	}

	ctx = context.WithValue(ctx, imageURLCtxKey, imageURL)
	ctx = context.WithValue(ctx, processingOptionsCtxKey, po)

	return ctx, nil
}

func getImageURL(ctx context.Context) string {
	str, _ := ctx.Value(imageURLCtxKey).(string)
	return str
}

func getProcessingOptions(ctx context.Context) *processingOptions {
	return ctx.Value(processingOptionsCtxKey).(*processingOptions)
}
