package imagedata

import (
	"bytes"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// NewFromBytes creates a new ImageData instance from the provided byte slice.
func NewFromBytes(b []byte, headers http.Header) (*ImageData, error) {
	r := bytes.NewReader(b)

	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	return NewFromBytesWithFormat(meta.Format(), b, headers), nil
}

// NewFromBytesWithFormat creates a new ImageData instance from the provided format and byte slice.
func NewFromBytesWithFormat(format imagetype.Type, b []byte, headers http.Header) *ImageData {
	return &ImageData{
		data:    b,
		format:  format,
		headers: headers,
	}
}
