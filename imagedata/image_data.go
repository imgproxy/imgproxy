package imagedata

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
)

var (
	Watermark     *ImageData
	FallbackImage *ImageData
)

type ImageData struct {
	format  imagetype.Type
	data    []byte
	Headers map[string]string

	cancel     context.CancelFunc
	cancelOnce sync.Once
}

func (d *ImageData) Close() error {
	d.cancelOnce.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})

	return nil
}

// Format returns the image format based on the metadata
func (d *ImageData) Format() imagetype.Type {
	return d.format
}

// Reader returns an io.ReadSeeker for the image data
func (d *ImageData) Reader() io.ReadSeeker {
	return bytes.NewReader(d.data)
}

// Size returns the size of the image data in bytes.
// NOTE: asyncbuffer implementation will .Wait() for the data to be fully read
func (d *ImageData) Size() (int, error) {
	return len(d.data), nil
}

func (d *ImageData) SetCancel(cancel context.CancelFunc) {
	d.cancel = cancel
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

func loadWatermark() (err error) {
	if len(config.WatermarkData) > 0 {
		Watermark, err = FromBase64(config.WatermarkData, "watermark", security.DefaultOptions())
		return
	}

	if len(config.WatermarkPath) > 0 {
		Watermark, err = FromFile(config.WatermarkPath, "watermark", security.DefaultOptions())
		return
	}

	if len(config.WatermarkURL) > 0 {
		Watermark, err = Download(context.Background(), config.WatermarkURL, "watermark", DownloadOptions{Header: nil, CookieJar: nil}, security.DefaultOptions())
		return
	}

	return nil
}

func loadFallbackImage() (err error) {
	switch {
	case len(config.FallbackImageData) > 0:
		FallbackImage, err = FromBase64(config.FallbackImageData, "fallback image", security.DefaultOptions())
	case len(config.FallbackImagePath) > 0:
		FallbackImage, err = FromFile(config.FallbackImagePath, "fallback image", security.DefaultOptions())
	case len(config.FallbackImageURL) > 0:
		FallbackImage, err = Download(context.Background(), config.FallbackImageURL, "fallback image", DownloadOptions{Header: nil, CookieJar: nil}, security.DefaultOptions())
	default:
		FallbackImage, err = nil, nil
	}

	if FallbackImage != nil && err == nil && config.FallbackImageTTL > 0 {
		if FallbackImage.Headers == nil {
			FallbackImage.Headers = make(map[string]string)
		}
		FallbackImage.Headers["Fallback-Image"] = "1"
	}

	return err
}

func FromBase64(encoded, desc string, secopts security.Options) (*ImageData, error) {
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
	size := 4 * (len(encoded)/3 + 1)

	imgdata, err := readAndCheckImage(dec, size, secopts)
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s: %s", desc, err)
	}

	return imgdata, nil
}

func FromFile(path, desc string, secopts security.Options) (*ImageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	imgdata, err := readAndCheckImage(f, int(fi.Size()), secopts)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	return imgdata, nil
}

func Download(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (*ImageData, error) {
	imgdata, err := download(ctx, imageURL, opts, secopts)
	if err != nil {
		return nil, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("Can't download %s", desc)),
		)
	}

	return imgdata, nil
}
