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
)

type urlOptions map[string][]string

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

type cropOptions struct {
	Width   int
	Height  int
	Gravity gravityOptions
}

type watermarkOptions struct {
	Enabled   bool
	Opacity   float64
	Replicate bool
	Gravity   gravityType
	OffsetX   int
	OffsetY   int
	Scale     float64
}

type processingOptions struct {
	Resize     resizeType
	Width      int
	Height     int
	Dpr        float64
	Gravity    gravityOptions
	Enlarge    bool
	Extend     bool
	Crop       cropOptions
	Format     imageType
	Quality    int
	Flatten    bool
	Background rgbColor
	Blur       float32
	Sharpen    float32

	CacheBuster string

	Watermark watermarkOptions

	PreferWebP  bool
	EnforceWebP bool

	UsedPresets []string
}

const (
	imageURLCtxKey          = ctxKey("imageUrl")
	processingOptionsCtxKey = ctxKey("processingOptions")
	urlTokenPlain           = "plain"
	maxClientHintDPR        = 8

	msgForbidden  = "Forbidden"
	msgInvalidURL = "Invalid URL"
)

var (
	errInvalidImageURL                    = errors.New("Invalid image url")
	errInvalidURLEncoding                 = errors.New("Invalid url encoding")
	errResultingImageFormatIsNotSupported = errors.New("Resulting image format is not supported")
	errInvalidPath                        = newError(404, "Invalid path", msgInvalidURL)
)

func (gt gravityType) String() string {
	for k, v := range gravityTypes {
		if v == gt {
			return k
		}
	}
	return ""
}

func (rt resizeType) String() string {
	for k, v := range resizeTypes {
		if v == rt {
			return k
		}
	}
	return ""
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

func decodeBase64URL(parts []string) (string, string, error) {
	var format string

	urlParts := strings.Split(strings.Join(parts, ""), ".")

	if len(urlParts) > 2 {
		return "", "", errInvalidURLEncoding
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	imageURL, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(urlParts[0], "="))
	if err != nil {
		return "", "", errInvalidURLEncoding
	}

	fullURL := fmt.Sprintf("%s%s", conf.BaseURL, string(imageURL))

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		return "", "", errInvalidImageURL
	}

	return fullURL, format, nil
}

func decodePlainURL(parts []string) (string, string, error) {
	var format string

	urlParts := strings.Split(strings.Join(parts, "/"), "@")

	if len(urlParts) > 2 {
		return "", "", errInvalidURLEncoding
	}

	if len(urlParts) == 2 && len(urlParts[1]) > 0 {
		format = urlParts[1]
	}

	if unescaped, err := url.PathUnescape(urlParts[0]); err == nil {
		fullURL := fmt.Sprintf("%s%s", conf.BaseURL, unescaped)
		if _, err := url.ParseRequestURI(fullURL); err == nil {
			return fullURL, format, nil
		}
	}

	return "", "", errInvalidImageURL
}

func decodeURL(parts []string) (string, string, error) {
	if len(parts) == 0 {
		return "", "", errInvalidURLEncoding
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

	po.Enlarge = args[0] != "0"

	return nil
}

func applyExtendOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid extend arguments: %v", args)
	}

	po.Extend = args[0] != "0"

	return nil
}

func applySizeOption(po *processingOptions, args []string) (err error) {
	if len(args) > 4 {
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

	if len(args) == 4 && len(args[3]) > 0 {
		if err = applyExtendOption(po, args[3:4]); err != nil {
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
		po.Resize = r
	} else {
		return fmt.Errorf("Invalid resize type: %s", args[0])
	}

	return nil
}

func applyResizeOption(po *processingOptions, args []string) error {
	if len(args) > 5 {
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

	if err := parseDimension(&po.Crop.Width, "crop width", args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		if err := parseDimension(&po.Crop.Height, "crop height", args[1]); err != nil {
			return err
		}
	}

	if len(args) > 2 {
		return parseGravity(&po.Crop.Gravity, args[2:])
	}

	return nil
}

func applyQualityOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid quality arguments: %v", args)
	}

	if q, err := strconv.Atoi(args[0]); err == nil && q > 0 && q <= 100 {
		po.Quality = q
	} else {
		return fmt.Errorf("Invalid quality: %s", args[0])
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
			return fmt.Errorf("Unknown asset: %s", preset)
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
			po.Watermark.Gravity = g
		} else {
			return fmt.Errorf("Invalid watermark position: %s", args[1])
		}
	}

	if len(args) > 2 && len(args[2]) > 0 {
		if x, err := strconv.Atoi(args[2]); err == nil {
			po.Watermark.OffsetX = x
		} else {
			return fmt.Errorf("Invalid watermark X offset: %s", args[2])
		}
	}

	if len(args) > 3 && len(args[3]) > 0 {
		if y, err := strconv.Atoi(args[3]); err == nil {
			po.Watermark.OffsetY = y
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

	if !vipsTypeSupportSave[po.Format] {
		return errResultingImageFormatIsNotSupported
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

func applyProcessingOption(po *processingOptions, name string, args []string) error {
	switch name {
	case "format", "f", "ext":
		if err := applyFormatOption(po, args); err != nil {
			return err
		}
	case "resize", "rs":
		if err := applyResizeOption(po, args); err != nil {
			return err
		}
	case "resizing_type", "rt":
		if err := applyResizingTypeOption(po, args); err != nil {
			return err
		}
	case "size", "s":
		if err := applySizeOption(po, args); err != nil {
			return err
		}
	case "width", "w":
		if err := applyWidthOption(po, args); err != nil {
			return err
		}
	case "height", "h":
		if err := applyHeightOption(po, args); err != nil {
			return err
		}
	case "enlarge", "el":
		if err := applyEnlargeOption(po, args); err != nil {
			return err
		}
	case "extend", "ex":
		if err := applyExtendOption(po, args); err != nil {
			return err
		}
	case "dpr":
		if err := applyDprOption(po, args); err != nil {
			return err
		}
	case "gravity", "g":
		if err := applyGravityOption(po, args); err != nil {
			return err
		}
	case "crop", "c":
		if err := applyCropOption(po, args); err != nil {
			return err
		}
	case "quality", "q":
		if err := applyQualityOption(po, args); err != nil {
			return err
		}
	case "background", "bg":
		if err := applyBackgroundOption(po, args); err != nil {
			return err
		}
	case "blur", "bl":
		if err := applyBlurOption(po, args); err != nil {
			return err
		}
	case "sharpen", "sh":
		if err := applySharpenOption(po, args); err != nil {
			return err
		}
	case "watermark", "wm":
		if err := applyWatermarkOption(po, args); err != nil {
			return err
		}
	case "preset", "pr":
		if err := applyPresetOption(po, args); err != nil {
			return err
		}
	case "cachebuster", "cb":
		if err := applyCacheBusterOption(po, args); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown processing option: %s", name)
	}

	return nil
}

func applyProcessingOptions(po *processingOptions, options urlOptions) error {
	for name, args := range options {
		if err := applyProcessingOption(po, name, args); err != nil {
			return err
		}
	}

	return nil
}

func parseURLOptions(opts []string) (urlOptions, []string) {
	parsed := make(urlOptions)
	urlStart := len(opts) + 1

	for i, opt := range opts {
		args := strings.Split(opt, ":")

		if len(args) == 1 {
			urlStart = i
			break
		}

		parsed[args[0]] = args[1:]
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
	var err error

	po := processingOptions{
		Resize:      resizeFit,
		Width:       0,
		Height:      0,
		Gravity:     gravityOptions{Type: gravityCenter},
		Enlarge:     false,
		Quality:     conf.Quality,
		Format:      imageTypeUnknown,
		Background:  rgbColor{255, 255, 255},
		Blur:        0,
		Sharpen:     0,
		Dpr:         1,
		Watermark:   watermarkOptions{Opacity: 1, Replicate: false, Gravity: gravityCenter},
		UsedPresets: make([]string, 0, len(conf.Presets)),
	}

	if strings.Contains(headers.Accept, "image/webp") {
		po.PreferWebP = conf.EnableWebpDetection || conf.EnforceWebp
		po.EnforceWebP = conf.EnforceWebp
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
		err = applyPresetOption(&po, []string{"default"})
	}

	return &po, err
}

func parsePathAdvanced(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	options, urlParts := parseURLOptions(parts)

	if err := applyProcessingOptions(po, options); err != nil {
		return "", po, err
	}

	url, extension, err := decodeURL(urlParts)
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

func parsePathPresets(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	presets := strings.Split(parts[0], ":")
	urlParts := parts[1:]

	if err := applyPresetOption(po, presets); err != nil {
		return "", nil, err
	}

	url, extension, err := decodeURL(urlParts)
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

func parsePathBasic(parts []string, headers *processingHeaders) (string, *processingOptions, error) {
	var err error

	if len(parts) < 6 {
		return "", nil, errInvalidPath
	}

	po, err := defaultProcessingOptions(headers)
	if err != nil {
		return "", po, err
	}

	po.Resize = resizeTypes[parts[0]]

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
	path := r.URL.RawPath
	if len(path) == 0 {
		path = r.URL.Path
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if len(parts) < 3 {
		return ctx, errInvalidPath
	}

	if !conf.AllowInsecure {
		if err := validatePath(parts[0], strings.TrimPrefix(path, fmt.Sprintf("/%s", parts[0]))); err != nil {
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
	var err error

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

	ctx = context.WithValue(ctx, imageURLCtxKey, imageURL)
	ctx = context.WithValue(ctx, processingOptionsCtxKey, po)

	return ctx, nil
}

func getImageURL(ctx context.Context) string {
	return ctx.Value(imageURLCtxKey).(string)
}

func getProcessingOptions(ctx context.Context) *processingOptions {
	return ctx.Value(processingOptionsCtxKey).(*processingOptions)
}
