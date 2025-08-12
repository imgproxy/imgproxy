package imagedata

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
)

// NewFromBytesWithFormat creates a new ImageData instance from the provided format
// and byte slice.
func NewFromBytesWithFormat(format imagetype.Type, b []byte) ImageData {
	return &imageDataBytes{
		data:   b,
		format: format,
		cancel: nil,
	}
}

// NewFromBytes creates a new ImageData instance from the provided byte slice.
func NewFromBytes(b []byte) (ImageData, error) {
	r := bytes.NewReader(b)

	format, err := imagetype.Detect(r)
	if err != nil {
		return nil, err
	}

	return NewFromBytesWithFormat(format, b), nil
}

// NewFromPath creates a new ImageData from an os.File
func NewFromPath(path string) (ImageData, error) {
	fl, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	b, err := io.ReadAll(fl)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(b)
}

// NewFromBase64 creates a new ImageData from a base64 encoded byte slice
func NewFromBase64(encoded string) (ImageData, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return NewFromBytes(b)
}

// sendRequest is a common logic between sync and async download.
func sendRequest(ctx context.Context, url string, opts DownloadOptions) (*imagefetcher.Request, *http.Response, http.Header, error) {
	h := make(http.Header)

	// NOTE: This will be removed in the future when our test context gets better isolation
	if len(redirectAllRequestsTo) > 0 {
		url = redirectAllRequestsTo
	}

	req, err := Fetcher.BuildRequest(ctx, url, opts.Header, opts.CookieJar)
	if err != nil {
		return req, nil, h, err
	}

	res, err := req.FetchImage()
	if res != nil {
		h = res.Header.Clone()
	}
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		req.Cancel()

		return req, nil, h, err
	}

	res, err = security.LimitResponseSize(res, opts.MaxSrcFileSize)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		req.Cancel()

		return req, nil, h, err
	}

	return req, res, h, nil
}

// DownloadSync downloads the image synchronously and returns the ImageData and HTTP headers.
func downloadSync(ctx context.Context, imageURL string, opts DownloadOptions) (ImageData, http.Header, error) {
	req, res, h, err := sendRequest(ctx, imageURL, opts)
	if res != nil {
		defer res.Body.Close()
	}

	if req != nil {
		defer req.Cancel()
	}

	if err != nil {
		return nil, h, err
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, h, err
	}

	format, err := imagetype.Detect(bytes.NewReader(b))
	if err != nil {
		return nil, h, err
	}

	d := NewFromBytesWithFormat(format, b)
	return d, h, err
}

// downloadAsync downloads the image asynchronously and returns the ImageData
// backed by AsyncBuffer and HTTP headers.
func downloadAsync(ctx context.Context, imageURL string, opts DownloadOptions) (ImageData, http.Header, error) {
	// We pass this responsibility to AsyncBuffer
	//nolint:bodyclose
	req, res, h, err := sendRequest(ctx, imageURL, opts)
	if err != nil {
		return nil, h, err
	}

	b := asyncbuffer.New(res.Body)

	format, err := imagetype.Detect(b.Reader())
	if err != nil {
		b.Close()
		req.Cancel()
		return nil, h, err
	}

	d := &imageDataAsyncBuffer{
		b:      b,
		format: format,
		cancel: nil,
	}
	d.AddCancel(req.Cancel) // request will be closed when the image data is consumed

	return d, h, err
}

// DownloadSyncWithDesc downloads the image synchronously and returns the ImageData, but
// wraps errors with desc.
func DownloadSync(ctx context.Context, imageURL, desc string, opts DownloadOptions) (ImageData, http.Header, error) {
	imgdata, h, err := downloadSync(ctx, imageURL, opts)
	if err != nil {
		return nil, h, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("can't download %s", desc)),
		)
	}

	return imgdata, h, nil
}

// DownloadSyncWithDesc downloads the image synchronously and returns the ImageData, but
// wraps errors with desc.
func DownloadAsync(ctx context.Context, imageURL, desc string, opts DownloadOptions) (ImageData, http.Header, error) {
	imgdata, h, err := downloadAsync(ctx, imageURL, opts)
	if err != nil {
		return nil, h, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("can't download %s", desc)),
		)
	}

	return imgdata, h, nil
}
