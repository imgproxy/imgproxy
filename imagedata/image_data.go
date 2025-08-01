package imagedata

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
)

var (
	Watermark     ImageData
	FallbackImage ImageData
)

type ImageData interface {
	io.Closer                     // Close closes the image data and releases any resources held by it
	Reader() io.ReadSeeker        // Reader returns a new ReadSeeker for the image data
	Format() imagetype.Type       // Format returns the image format from the metadata (shortcut)
	Size() (int, error)           // Size returns the size of the image data in bytes
	AddCancel(context.CancelFunc) // AddCancel attaches a cancel function to the image data

	// This will be removed in the future
	Headers() http.Header // Headers returns the HTTP headers of the image data, will be removed in the future
}

// imageDataBytes represents image data stored in a byte slice in memory
type imageDataBytes struct {
	format  imagetype.Type
	data    []byte
	headers http.Header

	cancel     []context.CancelFunc
	cancelOnce sync.Once
}

func (d *imageDataBytes) Close() error {
	d.cancelOnce.Do(func() {
		for _, cancel := range d.cancel {
			cancel()
		}
	})

	return nil
}

// Format returns the image format based on the metadata
func (d *imageDataBytes) Format() imagetype.Type {
	return d.format
}

// Reader returns an io.ReadSeeker for the image data
func (d *imageDataBytes) Reader() io.ReadSeeker {
	return bytes.NewReader(d.data)
}

// Size returns the size of the image data in bytes.
// NOTE: asyncbuffer implementation will .Wait() for the data to be fully read
func (d *imageDataBytes) Size() (int, error) {
	return len(d.data), nil
}

func (d *imageDataBytes) Headers() http.Header {
	return d.headers
}

func (d *imageDataBytes) AddCancel(cancel context.CancelFunc) {
	d.cancel = append(d.cancel, cancel)
}

func Init() error {
	initRead()

	if err := initDownloading(); err != nil {
		return err
	}

	if err := loadWatermark(); err != nil {
		return err
	}

	if err := loadFallbackImage(); err != nil {
		return err
	}

	return nil
}

func loadWatermark() error {
	var err error

	if len(config.WatermarkData) > 0 {
		Watermark, err = NewFromBase64(config.WatermarkData, security.DefaultOptions())

		// NOTE: this should be something like err = ierrors.Wrap(err).WithStackDeep(0).WithPrefix("watermark")
		// In the NewFromBase64 all errors should be wrapped to something like
		// .WithPrefix("load from base64")
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load watermark from Base64"))
		}
	}

	if len(config.WatermarkPath) > 0 {
		Watermark, err = NewFromPath(config.WatermarkPath, security.DefaultOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't read watermark from file"))
		}
	}

	if len(config.WatermarkURL) > 0 {
		Watermark, err = Download(context.Background(), config.WatermarkURL, "watermark", DownloadOptions{Header: nil, CookieJar: nil}, security.DefaultOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't download from URL"))
		}
	}

	return nil
}

func loadFallbackImage() (err error) {
	switch {
	case len(config.FallbackImageData) > 0:
		FallbackImage, err = NewFromBase64(config.FallbackImageData, security.DefaultOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load fallback image from Base64"))
		}

	case len(config.FallbackImagePath) > 0:
		FallbackImage, err = NewFromPath(config.FallbackImagePath, security.DefaultOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't read fallback image from file"))
		}

	case len(config.FallbackImageURL) > 0:
		FallbackImage, err = Download(context.Background(), config.FallbackImageURL, "fallback image", DownloadOptions{Header: nil, CookieJar: nil}, security.DefaultOptions())
	default:
		FallbackImage, err = nil, nil
	}

	if FallbackImage != nil && err == nil && config.FallbackImageTTL > 0 {
		FallbackImage.Headers().Set("Fallback-Image", "1")
	}

	return err
}

func Download(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (ImageData, error) {
	imgdata, err := download(ctx, imageURL, opts, secopts)
	if err != nil {
		return nil, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("Can't download %s", desc)),
		)
	}

	return imgdata, nil
}
