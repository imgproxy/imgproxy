package imagedatanew

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/security"
)

// NewFromFile creates a new ImageData from an os.File
func NewFromFile(path string, headers http.Header, secopts security.Options) (*imageDataBytes, error) {
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
	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		return nil, err
	}

	return &imageDataBytes{b, meta, headers.Clone()}, nil
}

// NewFromBase64 creates a new ImageData from a base64 encoded byte slice
func NewFromBase64(encoded string, headers http.Header, secopts security.Options) (*imageDataBytes, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(b)

	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		return nil, err
	}

	return &imageDataBytes{b, meta, headers.Clone()}, nil
}

// NewFromBytes creates a new ImageDataBytes from a byte slice
func NewFromBytes(b []byte, headers http.Header, secopts security.Options) (*imageDataBytes, error) {
	r := bytes.NewReader(b)

	meta, err := imagemeta.DecodeMeta(r)
	if err != nil {
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		return nil, err
	}

	return &imageDataBytes{b, meta, headers.Clone()}, nil
}

// NewFromResponse creates a new ImageDataResponse from an http.Response
func NewFromResponse(or *http.Response, secopts security.Options) (*imageDataResponse, error) {
	// We must not close the response body here, as is is read in background
	//nolint:bodyclose
	r, err := security.LimitResponseSize(or, secopts)
	if err != nil {
		return nil, err
	}

	b := asyncbuffer.FromReader(r.Body)
	c := r.Body

	meta, err := imagemeta.DecodeMeta(b.Reader())
	if err != nil {
		b.Close() // Close the async buffer early
		return nil, err
	}

	err = security.CheckMeta(meta, secopts)
	if err != nil {
		b.Close() // Close the async buffer early
		return nil, err
	}

	return &imageDataResponse{b, c, meta, or.Header.Clone()}, nil
}
