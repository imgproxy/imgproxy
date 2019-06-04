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
	}

	mimes = map[imageType]string{
		imageTypeJPEG: "image/jpeg",
		imageTypePNG:  "image/png",
		imageTypeWEBP: "image/webp",
		imageTypeGIF:  "image/gif",
		imageTypeICO:  "image/x-icon",
		imageTypeHEIC: "image/heif",
	}

	contentDispositionsFmt = map[imageType]string{
		imageTypeJPEG: "inline; filename=\"%s.jpg\"",
		imageTypePNG:  "inline; filename=\"%s.png\"",
		imageTypeWEBP: "inline; filename=\"%s.webp\"",
		imageTypeGIF:  "inline; filename=\"%s.gif\"",
		imageTypeICO:  "inline; filename=\"%s.ico\"",
		imageTypeHEIC: "inline; filename=\"%s.heic\"",
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

func (it imageType) Mime() string {
	if mime, ok := mimes[it]; ok {
		return mime
	}

	return "application/octet-stream"
}

func (it imageType) ContentDisposition(imageURL string) string {
	format, ok := contentDispositionsFmt[it]
	if !ok {
		return "inline"
	}

	url, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Sprintf(format, contentDispositionFilenameFallback)
	}

	_, filename := filepath.Split(url.Path)
	if len(filename) == 0 {
		return fmt.Sprintf(format, contentDispositionFilenameFallback)
	}

	return fmt.Sprintf(format, strings.TrimSuffix(filename, filepath.Ext(filename)))
}
