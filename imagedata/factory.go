package imagedata

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
)

// NewFromBytesWithFormat creates a new ImageData instance from the provided format,
// http headers and byte slice.
func NewFromBytesWithFormat(format imagetype.Type, b []byte, headers http.Header) ImageData {
	return &imageDataBytes{
		data:    b,
		format:  format,
		headers: headers,
		cancel:  make([]context.CancelFunc, 0),
	}
}

// NewFromBytes creates a new ImageData instance from the provided byte slice.
func NewFromBytes(b []byte) (ImageData, error) {
	r := bytes.NewReader(b)

	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	return NewFromBytesWithFormat(meta.Format(), b, make(http.Header)), nil
}

// NewFromPath creates a new ImageData from an os.File
func NewFromPath(path string, secopts security.Options) (ImageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fr, err := security.LimitFileSize(f, secopts)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	// NOTE: This will be removed in the future in favor of VIPS metadata extraction
	// It's here temporarily to maintain compatibility with existing code
	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(b)
}

// NewFromBase64 creates a new ImageData from a base64 encoded byte slice
func NewFromBase64(encoded string, secopts security.Options) (ImageData, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	// NOTE: This will be removed in the future in favor of VIPS metadata extraction
	// It's here temporarily to maintain compatibility with existing code
	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(b)
}
