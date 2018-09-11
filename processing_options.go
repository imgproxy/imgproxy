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
)

var gravityTypes = map[string]gravityType{
	"ce": gravityCenter,
	"no": gravityNorth,
	"ea": gravityEast,
	"so": gravitySouth,
	"we": gravityWest,
	"sm": gravitySmart,
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

type processingOptions struct {
	Resize  resizeType
	Width   int
	Height  int
	Gravity gravityType
	Enlarge bool
	Format  imageType
	Blur    float32
	Sharpen float32
}

func decodeURL(parts []string) (string, imageType, error) {
	var imgType imageType = imageTypeJPEG

	urlParts := strings.Split(strings.Join(parts, ""), ".")

	if len(urlParts) > 2 {
		return "", 0, errors.New("Invalid url encoding")
	}

	if len(urlParts) == 2 {
		if f, ok := imageTypes[urlParts[1]]; ok {
			imgType = f
		} else {
			return "", 0, fmt.Errorf("Invalid image format: %s", urlParts[1])
		}
	}

	url, err := base64.RawURLEncoding.DecodeString(urlParts[0])
	if err != nil {
		return "", 0, errors.New("Invalid url encoding")
	}

	return string(url), imgType, nil
}

func applyWidthOption(po *processingOptions, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("Invalid width arguments: %v", args)
	}

	if w, err := strconv.Atoi(args[0]); err == nil || w >= 0 {
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

	if h, err := strconv.Atoi(args[0]); err == nil || po.Height >= 0 {
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
	if len(args) > 1 {
		return fmt.Errorf("Invalid resize arguments: %v", args)
	}

	if g, ok := gravityTypes[args[0]]; ok {
		po.Gravity = g
	} else {
		return fmt.Errorf("Invalid gravity: %s", args[0])
	}

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
			for name, pargs := range p {
				if err := applyProcessingOption(po, name, pargs); err != nil {
					return err
				}
			}
		} else {
			return fmt.Errorf("Unknown asset: %s", preset)
		}
	}

	return nil
}

func applyFormatOption(po *processingOptions, imgType imageType) error {
	if !vipsTypeSupportSave[imgType] {
		return errors.New("Resulting image type not supported")
	}

	po.Format = imgType

	return nil
}

func applyProcessingOption(po *processingOptions, name string, args []string) error {
	switch name {
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

func defaultProcessingOptions() (processingOptions, error) {
	var err error

	po := processingOptions{
		Resize:  resizeFit,
		Width:   0,
		Height:  0,
		Gravity: gravityCenter,
		Enlarge: false,
		Format:  imageTypeJPEG,
		Blur:    0,
		Sharpen: 0,
	}

	if _, ok := conf.Presets["default"]; ok {
		err = applyPresetOption(&po, []string{"default"})
	}

	return po, err
}

func parsePathAdvanced(parts []string) (string, processingOptions, error) {
	po, err := defaultProcessingOptions()
	if err != nil {
		return "", po, err
	}

	options, urlParts := parseURLOptions(parts)

	for name, args := range options {
		if err := applyProcessingOption(&po, name, args); err != nil {
			return "", po, err
		}
	}

	url, imgType, err := decodeURL(urlParts)
	if err != nil {
		return "", po, err
	}

	if err := applyFormatOption(&po, imgType); err != nil {
		return "", po, errors.New("Resulting image type not supported")
	}

	return string(url), po, nil
}

func parsePathSimple(parts []string) (string, processingOptions, error) {
	var err error

	if len(parts) < 6 {
		return "", processingOptions{}, errors.New("Invalid path")
	}

	po, err := defaultProcessingOptions()
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

	if err = applyGravityOption(&po, parts[3:4]); err != nil {
		return "", po, err
	}

	if err = applyEnlargeOption(&po, parts[4:5]); err != nil {
		return "", po, err
	}

	url, imgType, err := decodeURL(parts[5:])
	if err != nil {
		return "", po, err
	}

	if err := applyFormatOption(&po, imgType); err != nil {
		return "", po, errors.New("Resulting image type not supported")
	}

	return string(url), po, nil
}

func parsePath(r *http.Request) (string, processingOptions, error) {
	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if len(parts) < 3 {
		return "", processingOptions{}, errors.New("Invalid path")
	}

	if !conf.AllowInsecure {
		if err := validatePath(parts[0], strings.TrimPrefix(path, fmt.Sprintf("/%s", parts[0]))); err != nil {
			return "", processingOptions{}, err
		}
	}

	if _, ok := resizeTypes[parts[1]]; ok {
		return parsePathSimple(parts[1:])
	} else {
		return parsePathAdvanced(parts[1:])
	}
}
