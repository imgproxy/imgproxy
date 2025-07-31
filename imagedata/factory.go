package imagedata

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/imgproxy/imgproxy/v3/imagemeta"
)

// NewFromBytes creates a new ImageData instance from the provided byte slice.
func NewFromBytes(b []byte, headers http.Header) (*ImageData, error) {
	r := bytes.NewReader(b)

	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	// Temporary workaround for the old ImageData interface
	h := make(map[string]string, len(headers))
	for k, v := range headers {
		h[k] = strings.Join(v, ", ")
	}

	return &ImageData{
		data:    b,
		meta:    meta,
		Headers: h,
	}, nil
}
