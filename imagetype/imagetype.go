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
		JPEG: "%s; filename=\"%s.jpg\"",
		PNG:  "%s; filename=\"%s.png\"",
		WEBP: "%s; filename=\"%s.webp\"",
		GIF:  "%s; filename=\"%s.gif\"",
		ICO:  "%s; filename=\"%s.ico\"",
		SVG:  "%s; filename=\"%s.svg\"",
		HEIC: "%s; filename=\"%s.heic\"",
		AVIF: "%s; filename=\"%s.avif\"",
		BMP:  "%s; filename=\"%s.bmp\"",
		TIFF: "%s; filename=\"%s.tiff\"",
	}
)

func ByMime(mime string) Type {
	for k, v := range mimes {
		if v == mime {
			return k
		}
	}
	return Unknown
}

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

func (it Type) ContentDisposition(filename string, returnAttachment bool) string {
	disposition := "inline"

	if returnAttachment {
		disposition = "attachment"
	}

	format, ok := contentDispositionsFmt[it]
	if !ok {
		return disposition
	}

	return fmt.Sprintf(format, disposition, strings.ReplaceAll(filename, `"`, "%22"))
}

func (it Type) ContentDispositionFromURL(imageURL string, returnAttachment bool) string {
	url, err := url.Parse(imageURL)
	if err != nil {
		return it.ContentDisposition(contentDispositionFilenameFallback, returnAttachment)
	}

	_, filename := filepath.Split(url.Path)
	if len(filename) == 0 {
		return it.ContentDisposition(contentDispositionFilenameFallback, returnAttachment)
	}

	return it.ContentDisposition(strings.TrimSuffix(filename, filepath.Ext(filename)), returnAttachment)
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

func (it Type) SupportsThumbnail() bool {
	return it == HEIC || it == AVIF
}
