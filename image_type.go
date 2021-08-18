package main

/*
#cgo LDFLAGS: -s -w
#include "vips.h"
*/
import "C"

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type imageType int

const (
	imageTypeUnknown = imageType(C.UNKNOWN)
	imageTypeJPEG    = imageType(C.JPEG)
	imageTypePNG     = imageType(C.PNG)
	imageTypeWEBP    = imageType(C.WEBP)
	imageTypeGIF     = imageType(C.GIF)
	imageTypeICO     = imageType(C.ICO)
	imageTypeSVG     = imageType(C.SVG)
	imageTypeHEIC    = imageType(C.HEIC)
	imageTypeAVIF    = imageType(C.AVIF)
	imageTypeBMP     = imageType(C.BMP)
	imageTypeTIFF    = imageType(C.TIFF)

	contentDispositionFilenameFallback = "image"
)

var (
	imageTypes = map[string]imageType{
		"jpeg": imageTypeJPEG,
		"jpg":  imageTypeJPEG,
		"png":  imageTypePNG,
		"webp": imageTypeWEBP,
		"gif":  imageTypeGIF,
		"ico":  imageTypeICO,
		"svg":  imageTypeSVG,
		"heic": imageTypeHEIC,
		"avif": imageTypeAVIF,
		"bmp":  imageTypeBMP,
		"tiff": imageTypeTIFF,
	}

	mimes = map[imageType]string{
		imageTypeJPEG: "image/jpeg",
		imageTypePNG:  "image/png",
		imageTypeWEBP: "image/webp",
		imageTypeGIF:  "image/gif",
		imageTypeICO:  "image/x-icon",
		imageTypeSVG:  "image/svg+xml",
		imageTypeHEIC: "image/heif",
		imageTypeAVIF: "image/avif",
		imageTypeBMP:  "image/bmp",
		imageTypeTIFF: "image/tiff",
	}

	contentDispositionsFmt = map[imageType]string{
		imageTypeJPEG: "inline; filename=\"%s.jpg\"",
		imageTypePNG:  "inline; filename=\"%s.png\"",
		imageTypeWEBP: "inline; filename=\"%s.webp\"",
		imageTypeGIF:  "inline; filename=\"%s.gif\"",
		imageTypeICO:  "inline; filename=\"%s.ico\"",
		imageTypeSVG:  "inline; filename=\"%s.svg\"",
		imageTypeHEIC: "inline; filename=\"%s.heic\"",
		imageTypeAVIF: "inline; filename=\"%s.avif\"",
		imageTypeBMP:  "inline; filename=\"%s.bmp\"",
		imageTypeTIFF: "inline; filename=\"%s.tiff\"",
	}
)

func (it imageType) String() string {
	for k, v := range imageTypes {
		if v == it {
			return k
		}
	}
	return ""
}

func (it imageType) MarshalJSON() ([]byte, error) {
	for k, v := range imageTypes {
		if v == it {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}

func (it imageType) Mime() string {
	if mime, ok := mimes[it]; ok {
		return mime
	}

	return "application/octet-stream"
}

func (it imageType) ContentDisposition(filename string) string {
	format, ok := contentDispositionsFmt[it]
	if !ok {
		return "inline"
	}

	return fmt.Sprintf(format, filename)
}

func (it imageType) ContentDispositionFromURL(imageURL string) string {
	url, err := url.Parse(imageURL)
	if err != nil {
		return it.ContentDisposition(contentDispositionFilenameFallback)
	}

	_, filename := filepath.Split(url.Path)
	if len(filename) == 0 {
		return it.ContentDisposition(contentDispositionFilenameFallback)
	}

	return it.ContentDisposition(strings.TrimSuffix(filename, filepath.Ext(filename)))
}

func (it imageType) SupportsAlpha() bool {
	return it != imageTypeJPEG && it != imageTypeBMP
}

func (it imageType) SupportsColourProfile() bool {
	return it == imageTypeJPEG ||
		it == imageTypePNG ||
		it == imageTypeWEBP ||
		it == imageTypeAVIF
}
