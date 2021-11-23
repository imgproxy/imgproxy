package imagetype

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type Type int

const (
	Unknown Type = iota
	JPEG
	PNG
	WEBP
	GIF
	ICO
	SVG
	HEIC
	AVIF
	BMP
	TIFF
)

const contentDispositionFilenameFallback = "image"

var (
	Types = map[string]Type{
		"jpeg": JPEG,
		"jpg":  JPEG,
		"png":  PNG,
		"webp": WEBP,
		"gif":  GIF,
		"ico":  ICO,
		"svg":  SVG,
		"heic": HEIC,
		"avif": AVIF,
		"bmp":  BMP,
		"tiff": TIFF,
	}

	mimes = map[Type]string{
		JPEG: "image/jpeg",
		PNG:  "image/png",
		WEBP: "image/webp",
		GIF:  "image/gif",
		ICO:  "image/x-icon",
		SVG:  "image/svg+xml",
		HEIC: "image/heif",
		AVIF: "image/avif",
		BMP:  "image/bmp",
		TIFF: "image/tiff",
	}

	contentDispositionsFmt = map[Type]string{
		JPEG: "inline; filename=\"%s.jpg\"",
		PNG:  "inline; filename=\"%s.png\"",
		WEBP: "inline; filename=\"%s.webp\"",
		GIF:  "inline; filename=\"%s.gif\"",
		ICO:  "inline; filename=\"%s.ico\"",
		SVG:  "inline; filename=\"%s.svg\"",
		HEIC: "inline; filename=\"%s.heic\"",
		AVIF: "inline; filename=\"%s.avif\"",
		BMP:  "inline; filename=\"%s.bmp\"",
		TIFF: "inline; filename=\"%s.tiff\"",
	}
)

func (it Type) String() string {
	for k, v := range Types {
		if v == it {
			return k
		}
	}
	return ""
}

func (it Type) MarshalJSON() ([]byte, error) {
	for k, v := range Types {
		if v == it {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}

func (it Type) Mime() string {
	if mime, ok := mimes[it]; ok {
		return mime
	}

	return "application/octet-stream"
}

func (it Type) ContentDisposition(filename string) string {
	format, ok := contentDispositionsFmt[it]
	if !ok {
		return "inline"
	}

	return fmt.Sprintf(format, strings.ReplaceAll(filename, `"`, "%22"))
}

func (it Type) ContentDispositionFromURL(imageURL string) string {
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

func (it Type) SupportsAlpha() bool {
	return it != JPEG && it != BMP
}

func (it Type) SupportsAnimation() bool {
	return it == GIF || it == WEBP
}

func (it Type) SupportsColourProfile() bool {
	return it == JPEG ||
		it == PNG ||
		it == WEBP ||
		it == AVIF
}
