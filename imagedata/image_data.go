package imagedata

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	Watermark            ImageData
	FallbackImage        ImageData
	FallbackImageHeaders http.Header // Headers for the fallback image
)

// ImageData represents the data of an image that can be read from a source.
// Please note that this interface can be backed by any reader, including lazy AsyncBuffer.
// There is no other way to guarantee that the data is read without errors except reading it till EOF.
type ImageData interface {
	io.Closer                     // Close closes the image data and releases any resources held by it
	Reader() io.ReadSeeker        // Reader returns a new ReadSeeker for the image data
	Format() imagetype.Type       // Format returns the image format from the metadata (shortcut)
	Size() (int, error)           // Size returns the size of the image data in bytes
	AddCancel(context.CancelFunc) // AddCancel attaches a cancel function to the image data
	Error() error                 // Error returns any error that occurred during reading data from source
}

// imageDataBytes represents image data stored in a byte slice in memory
type imageDataBytes struct {
	format     imagetype.Type
	data       []byte
	cancel     []context.CancelFunc
	cancelOnce sync.Once
}

// imageDataAsyncBuffer is a struct that implements the ImageData interface backed by an AsyncBuffer
type imageDataAsyncBuffer struct {
	b          *asyncbuffer.AsyncBuffer
	format     imagetype.Type
	cancel     []context.CancelFunc
	cancelOnce sync.Once
}

// Close closes the image data and releases any resources held by it
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
func (d *imageDataBytes) Size() (int, error) {
	return len(d.data), nil
}

// AddCancel attaches a cancel function to the image data
func (d *imageDataBytes) AddCancel(cancel context.CancelFunc) {
	d.cancel = append(d.cancel, cancel)
}

func (d *imageDataBytes) Error() error {
	// No error handling for in-memory data, return nil
	return nil
}

// Reader returns a ReadSeeker for the image data
func (d *imageDataAsyncBuffer) Reader() io.ReadSeeker {
	return d.b.Reader()
}

// Close closes the response body (hence, response) and the async buffer itself
func (d *imageDataAsyncBuffer) Close() error {
	d.cancelOnce.Do(func() {
		d.b.Close()
		for _, cancel := range d.cancel {
			cancel()
		}
	})

	return nil
}

// Format returns the image format from the metadata
func (d *imageDataAsyncBuffer) Format() imagetype.Type {
	return d.format
}

// Size returns the size of the image data in bytes.
// It waits for the async buffer to finish reading.
func (d *imageDataAsyncBuffer) Size() (int, error) {
	return d.b.Wait()
}

// AddCancel attaches a cancel function to the image data
func (d *imageDataAsyncBuffer) AddCancel(cancel context.CancelFunc) {
	d.cancel = append(d.cancel, cancel)
}

// Error returns any error that occurred during reading data from
// async buffer or the underlying source.
func (d *imageDataAsyncBuffer) Error() error {
	return d.b.Error()
}

func Init() error {
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

	switch {
	case len(config.WatermarkData) > 0:
		Watermark, err = NewFromBase64(config.WatermarkData)

		// NOTE: this should be something like err = ierrors.Wrap(err).WithStackDeep(0).WithPrefix("watermark")
		// In the NewFromBase64 all errors should be wrapped to something like
		// .WithPrefix("load from base64")
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load watermark from Base64"))
		}

	case len(config.WatermarkPath) > 0:
		Watermark, err = NewFromPath(config.WatermarkPath)
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't read watermark from file"))
		}

	case len(config.WatermarkURL) > 0:
		Watermark, _, err = DownloadSync(context.Background(), config.WatermarkURL, "watermark", DefaultDownloadOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't download from URL"))
		}

	default:
		Watermark = nil
	}

	return nil
}

func loadFallbackImage() (err error) {
	switch {
	case len(config.FallbackImageData) > 0:
		FallbackImage, err = NewFromBase64(config.FallbackImageData)
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load fallback image from Base64"))
		}

	case len(config.FallbackImagePath) > 0:
		FallbackImage, err = NewFromPath(config.FallbackImagePath)
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't read fallback image from file"))
		}

	case len(config.FallbackImageURL) > 0:
		FallbackImage, FallbackImageHeaders, err = DownloadSync(context.Background(), config.FallbackImageURL, "fallback image", DefaultDownloadOptions())
		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't download from URL"))
		}

	default:
		FallbackImage = nil
	}

	return err
}
