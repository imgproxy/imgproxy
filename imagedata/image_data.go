package imagedata

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	Watermark ImageData
)

// ImageData represents the data of an image that can be read from a source.
// Please note that this interface can be backed by any reader, including lazy AsyncBuffer.
// There is no other way to guarantee that the data is read without errors except reading it till EOF.
type ImageData interface {
	io.Closer               // Close closes the image data and releases any resources held by it
	Reader() io.ReadSeeker  // Reader returns a new ReadSeeker for the image data
	Format() imagetype.Type // Format returns the image format from the metadata (shortcut)
	Size() (int, error)     // Size returns the size of the image data in bytes
	Error() error           // Error returns any error that occurred during reading data from source

	// AddCancel attaches a cancel function to the image data.
	// Please note that Cancel functions must be idempotent: for instance, an implementation
	// could wrap cancel into sync.Once.
	AddCancel(cancel context.CancelFunc)
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
	desc       string
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
	if err := d.b.Error(); err != nil {
		return wrapDownloadError(err, d.desc)
	}
	return nil
}
