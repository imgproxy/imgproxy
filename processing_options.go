package main

/*
#cgo LDFLAGS: -s -w
#include "image_types.h"
*/
import "C"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type urlOptions map[string][]string

type imageType int

const (
	imageTypeUnknown = C.UNKNOWN
	imageTypeJPEG    = C.JPEG
	imageTypePNG     = C.PNG
	imageTypeWEBP    = C.WEBP
	imageTypeGIF     = C.GIF
)

var imageTypes = map[string]imageType{
	"jpeg": imageTypeJPEG,
	"jpg":  imageTypeJPEG,
	"png":  imageTypePNG,
	"webp": imageTypeWEBP,
	"gif":  imageTypeGIF,
}

type gravityType int

const (
	gravityCenter gravityType = iota
	gravityNorth
	gravityEast
	gravitySouth
	gravityWest
	gravitySmart
	gravityFocusPoint
)

var gravityTypes = map[string]gravityType{
	"ce": gravityCenter,
	"no": gravityNorth,
	"ea": gravityEast,
	"so": gravitySouth,
	"we": gravityWest,
	"sm": gravitySmart,
	"fp": gravityFocusPoint,
}

type gravity struct {
	Type gravityType
	X, Y float64
}

type resizeType int

const (
	resizeFit resizeType = iota
	resizeFill
	resizeCrop
)

var resizeTypes = map[string]resizeType{
	"fit":  resizeFit,
	"fill": resizeFill,
	"crop": resizeCrop,
}

type color struct{ R, G, B uint8 }

type processingOptions struct {
	Resize     resizeType
	Width      int
	Height     int
	Gravity    gravity
	Enlarge    bool
	Format     imageType
	Flatten    bool
	Background color
	Blur       float32
	Sharpen    float32
}

func (it imageType) String() string {
	for k, v := range imageTypes {
		if v == it {
			return k
		}
	}
	return ""
}

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

func decodeURL(parts []string) (string, string, error) {
	var extension string

	urlParts := strings.Split(strings.Join(parts, ""), ".")

	if len(urlParts) > 2 {
		return "", "", errors.New("Invalid url encoding")
	}

	if len(urlParts) == 2 {
		extension = urlParts[1]
	}

	url, err := base64.RawURLEncoding.DecodeString(urlParts[0])
	if err != nil {
		return "", "", errors.New("Invalid url encoding")
	}

	return string(url), extension, nil
}

func applyWidthOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid width arguments: %v", args)
	}

	if w, err := strconv.Atoi(args[0]); err == nil && w >= 0 {
		po.Width = w
	} else {
		return fmt.Errorf("Invalid width: %s", args[0])
	}

	return nil
}

func applyHeightOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid height arguments: %v", args)
	}

	if h, err := strconv.Atoi(args[0]); err == nil && po.Height >= 0 {
		po.Height = h
	} else {
		return fmt.Errorf("Invalid height: %s", args[0])
	}

	return nil
}

func applyEnlargeOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid enlarge arguments: %v", args)
	}

	po.Enlarge = args[0] != "0"

	return nil
}

func applySizeOption(po *processingOptions, args []string) (err error) {
	if len(args) > 3 {
		return fmt.Errorf("Invalid size arguments: %v", args)
	}

	if len(args) >= 1 {
		if err = applyWidthOption(po, args[0:1]); err != nil {
			return
		}
	}

	if len(args) >= 2 {
		if err = applyHeightOption(po, args[1:2]); err != nil {
			return
		}
	}

	if len(args) == 3 {
		if err = applyEnlargeOption(po, args[2:3]); err != nil {
			return
		}
	}

	return nil
}

func applyResizeOption(po *processingOptions, args []string) error {
	if len(args) > 4 {
		return fmt.Errorf("Invalid resize arguments: %v", args)
	}

	if r, ok := resizeTypes[args[0]]; ok {
		po.Resize = r
	} else {
		return fmt.Errorf("Invalid resize type: %s", args[0])
	}

	if len(args) > 1 {
		if err := applySizeOption(po, args[1:]); err != nil {
			return err
		}
	}

	return nil
}

func applyGravityOption(po *processingOptions, args []string) error {
	if g, ok := gravityTypes[args[0]]; ok {
		po.Gravity.Type = g
	} else {
		return fmt.Errorf("Invalid gravity: %s", args[0])
	}

	if po.Gravity.Type == gravityFocusPoint {
		if len(args) != 3 {
			return fmt.Errorf("Invalid gravity arguments: %v", args)
		}

		if x, err := strconv.ParseFloat(args[1], 64); err == nil && x >= 0 && x <= 1 {
			po.Gravity.X = x
		} else {
			return fmt.Errorf("Invalid gravity X: %s", args[1])
		}

		if y, err := strconv.ParseFloat(args[2], 64); err == nil && y >= 0 && y <= 1 {
			po.Gravity.Y = y
		} else {
			return fmt.Errorf("Invalid gravity Y: %s", args[2])
		}
	} else if len(args) > 1 {
		return fmt.Errorf("Invalid gravity arguments: %v", args)
	}

	return nil
}

func applyBackgroundOption(po *processingOptions, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("Invalid background arguments: %v", args)
	}

	if r, err := strconv.ParseUint(args[0], 10, 8); err == nil && r >= 0 && r <= 255 {
		po.Background.R = uint8(r)
	} else {
		return fmt.Errorf("Invalid background red channel: %s", args[1])
	}

	if g, err := strconv.ParseUint(args[1], 10, 8); err == nil && g >= 0 && g <= 255 {
		po.Background.G = uint8(g)
	} else {
		return fmt.Errorf("Invalid background green channel: %s", args[1])
	}

	if b, err := strconv.ParseUint(args[2], 10, 8); err == nil && b >= 0 && b <= 255 {
		po.Background.B = uint8(b)
	} else {
		return fmt.Errorf("Invalid background blue channel: %s", args[2])
	}

	po.Flatten = true

	return nil
}

func applyBlurOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid blur arguments: %v", args)
	}

	if b, err := strconv.ParseFloat(args[0], 32); err == nil || b >= 0 {
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

	if s, err := strconv.ParseFloat(args[0], 32); err == nil || s >= 0 {
		po.Sharpen = float32(s)
	} else {
		return fmt.Errorf("Invalid sharpen: %s", args[0])
	}

	return nil
}

func applyPresetOption(po *processingOptions, args []string) error {
	for _, preset := range args {
		if p, ok := conf.Presets[preset]; ok {
			if err := applyProcessingOptions(po, p); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Unknown asset: %s", preset)
		}
	}

	return nil
}

func applyFormatOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid format arguments: %v", args)
	}

	if conf.EnforceWebp && po.Format == imageTypeWEBP {
		// Webp is enforced and already set as format
		return nil
	}

	if f, ok := imageTypes[args[0]]; ok {
		po.Format = f
	} else {
		return fmt.Errorf("Invalid image format: %s", args[0])
	}

	if !vipsTypeSupportSave[po.Format] {
		return errors.New("Resulting image type not supported")
	}

	return nil
}

func applyProcessingOption(po *processingOptions, name string, args []string) error {
	switch name {
	case "format":
		if err := applyFormatOption(po, args); err != nil {
			return err
		}
	case "resize":
		if err := applyResizeOption(po, args); err != nil {
			return err
		}
	case "size":
		if err := applySizeOption(po, args); err != nil {
			return err
		}
	case "width":
		if err := applyWidthOption(po, args); err != nil {
			return err
		}
	case "height":
		if err := applyHeightOption(po, args); err != nil {
			return err
		}
	case "enlarge":
		if err := applyEnlargeOption(po, args); err != nil {
			return err
		}
	case "gravity":
		if err := applyGravityOption(po, args); err != nil {
			return err
		}
	case "background":
		if err := applyBackgroundOption(po, args); err != nil {
			return err
		}
	case "blur":
		if err := applyBlurOption(po, args); err != nil {
			return err
		}
	case "sharpen":
		if err := applySharpenOption(po, args); err != nil {
			return err
		}
	case "preset":
		if err := applyPresetOption(po, args); err != nil {
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

func defaultProcessingOptions(acceptHeader string) (processingOptions, error) {
	var err error

	po := processingOptions{
		Resize:  resizeFit,
		Width:   0,
		Height:  0,
		Gravity: gravity{Type: gravityCenter},
		Enlarge: false,
		Format:  imageTypeJPEG,
		Blur:    0,
		Sharpen: 0,
	}

	if (conf.EnableWebpDetection || conf.EnforceWebp) && strings.Contains(acceptHeader, "image/webp") {
		po.Format = imageTypeWEBP
	}

	if _, ok := conf.Presets["default"]; ok {
		err = applyPresetOption(&po, []string{"default"})
	}

	return po, err
}

func parsePathAdvanced(parts []string, acceptHeader string) (string, processingOptions, error) {
	po, err := defaultProcessingOptions(acceptHeader)
	if err != nil {
		return "", po, err
	}

	options, urlParts := parseURLOptions(parts)

	if err := applyProcessingOptions(&po, options); err != nil {
		return "", po, err
	}

	url, extension, err := decodeURL(urlParts)
	if err != nil {
		return "", po, err
	}

	if len(extension) > 0 {
		if err := applyFormatOption(&po, []string{extension}); err != nil {
			return "", po, errors.New("Resulting image type not supported")
		}
	}

	return string(url), po, nil
}

func parsePathSimple(parts []string, acceptHeader string) (string, processingOptions, error) {
	var err error

	if len(parts) < 6 {
		return "", processingOptions{}, errors.New("Invalid path")
	}

	po, err := defaultProcessingOptions(acceptHeader)
	if err != nil {
		return "", po, err
	}

	po.Resize = resizeTypes[parts[0]]

	if err = applyWidthOption(&po, parts[1:2]); err != nil {
		return "", po, err
	}

	if err = applyHeightOption(&po, parts[2:3]); err != nil {
		return "", po, err
	}

	if err = applyGravityOption(&po, strings.Split(parts[3], ":")); err != nil {
		return "", po, err
	}

	if err = applyEnlargeOption(&po, parts[4:5]); err != nil {
		return "", po, err
	}

	url, extension, err := decodeURL(parts[5:])
	if err != nil {
		return "", po, err
	}

	if len(extension) > 0 {
		if err := applyFormatOption(&po, []string{extension}); err != nil {
			return "", po, errors.New("Resulting image type not supported")
		}
	}

	return string(url), po, nil
}

func parsePath(r *http.Request) (string, processingOptions, error) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	var acceptHeader string
	if h, ok := r.Header["Accept"]; ok {
		acceptHeader = h[0]
	}

	if len(parts) < 3 {
		return "", processingOptions{}, errors.New("Invalid path")
	}

	if !conf.AllowInsecure {
		if err := validatePath(parts[0], strings.TrimPrefix(path, fmt.Sprintf("/%s", parts[0]))); err != nil {
			return "", processingOptions{}, err
		}
	}

	if _, ok := resizeTypes[parts[1]]; ok {
		return parsePathSimple(parts[1:], acceptHeader)
	}

	return parsePathAdvanced(parts[1:], acceptHeader)
}
