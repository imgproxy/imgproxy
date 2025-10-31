package imagedata

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

// Factory represents ImageData factory
type Factory struct {
	fetcher *fetcher.Fetcher
}

// NewFactory creates a new factory
func NewFactory(fetcher *fetcher.Fetcher) *Factory {
	return &Factory{
		fetcher: fetcher,
	}
}

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
func (f *Factory) NewFromBytes(b []byte) (ImageData, error) {
	r := bytes.NewReader(b)

	format, err := imagetype.Detect(r, "", "")
	if err != nil {
		return nil, err
	}

	return NewFromBytesWithFormat(format, b), nil
}

// NewFromPath creates a new ImageData from an os.File
func (f *Factory) NewFromPath(path string) (ImageData, error) {
	fl, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	b, err := io.ReadAll(fl)
	if err != nil {
		return nil, err
	}

	return f.NewFromBytes(b)
}

// NewFromBase64 creates a new ImageData from a base64 encoded byte slice
func (f *Factory) NewFromBase64(encoded string) (ImageData, error) {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return f.NewFromBytes(b)
}

// sendRequest is a common logic between sync and async download.
func (f *Factory) sendRequest(ctx context.Context, url string, opts DownloadOptions) (*fetcher.Request, *http.Response, http.Header, error) {
	h := make(http.Header)

	// NOTE: This will be removed in the future when our test context gets better isolation
	if len(redirectAllRequestsTo) > 0 {
		url = redirectAllRequestsTo
	}

	req, err := f.fetcher.BuildRequest(ctx, url, opts.Header, opts.CookieJar)
	if err != nil {
		return req, nil, h, err
	}

	res, err := req.Fetch()
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

	res, err = limitResponseSize(res, opts.MaxSrcFileSize)
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
func (f *Factory) DownloadSync(ctx context.Context, imageURL, desc string, opts DownloadOptions) (ImageData, http.Header, error) {
	if opts.DownloadFinished != nil {
		defer opts.DownloadFinished()
	}

	req, res, h, err := f.sendRequest(ctx, imageURL, opts)
	if res != nil {
		defer res.Body.Close()
	}

	if req != nil {
		defer req.Cancel()
	}

	if err != nil {
		return nil, h, wrapDownloadError(err, desc)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, h, wrapDownloadError(err, desc)
	}

	ct := res.Header.Get(httpheaders.ContentType)
	ext := strings.ToLower(filepath.Ext(res.Request.URL.Path))

	format, err := imagetype.Detect(bytes.NewReader(b), ct, ext)
	if err != nil {
		return nil, h, wrapDownloadError(err, desc)
	}

	d := NewFromBytesWithFormat(format, b)
	return d, h, nil
}

// DownloadAsync downloads the image asynchronously and returns the ImageData
// backed by AsyncBuffer and HTTP headers.
func (f *Factory) DownloadAsync(ctx context.Context, imageURL, desc string, opts DownloadOptions) (ImageData, http.Header, error) {
	// We pass this responsibility to AsyncBuffer
	//nolint:bodyclose
	req, res, h, err := f.sendRequest(ctx, imageURL, opts)
	if err != nil {
		if opts.DownloadFinished != nil {
			defer opts.DownloadFinished()
		}
		return nil, h, wrapDownloadError(err, desc)
	}

	b := asyncbuffer.New(res.Body, int(res.ContentLength), opts.DownloadFinished)

	ct := res.Header.Get(httpheaders.ContentType)
	ext := strings.ToLower(filepath.Ext(res.Request.URL.Path))

	format, err := imagetype.Detect(b.Reader(), ct, ext)
	if err != nil {
		b.Close()
		req.Cancel()
		return nil, h, wrapDownloadError(err, desc)
	}

	// We successfully detected the image type, so we can release the pause
	// and let the buffer read the rest of the data immediately.
	b.ReleaseThreshold()

	d := &imageDataAsyncBuffer{
		b:      b,
		format: format,
		desc:   desc,
		cancel: nil,
	}
	d.AddCancel(req.Cancel) // request will be closed when the image data is consumed

	return d, h, nil
}
