package imagedata

import (
	"bytes"
	"io"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
)

// Provider represents the data of an image that can be read from a source.
// Please note that this interface can be backed by any reader, including lazy AsyncBuffer.
// There is no other way to guarantee that the data is read without errors except reading it till EOF.
type Provider interface {
	Close() error          // Close releases any resources held by the provider
	Reader() io.ReadSeeker // Reader returns a new ReadSeeker for the image data
	Size() (int, error)    // Size returns the size of the image data in bytes
	Error() error          // Error returns any error that occurred during reading data from source
}

// bytesProvider represents image data stored in a byte slice in memory
type bytesProvider struct {
	data []byte
}

// asyncBufferProvider is a struct that implements the ImageData interface backed by an AsyncBuffer
type asyncBufferProvider struct {
	b    *asyncbuffer.AsyncBuffer
	desc string
}

// Reader returns an io.ReadSeeker for the image data
func (d *bytesProvider) Reader() io.ReadSeeker {
	return bytes.NewReader(d.data)
}

// Size returns the size of the image data in bytes.
func (d *bytesProvider) Size() (int, error) {
	return len(d.data), nil
}

// Error returns any error that occurred during reading data from source.
func (d *bytesProvider) Error() error {
	// No error handling for in-memory data, return nil
	return nil
}

// Close no close for in-memory data, return nil
func (d *bytesProvider) Close() error {
	// No resources to release for in-memory data, return nil
	return nil
}

// Reader returns a ReadSeeker for the image data
func (d *asyncBufferProvider) Reader() io.ReadSeeker {
	return d.b.Reader()
}

// Size returns the size of the image data in bytes.
// It waits for the async buffer to finish reading.
func (d *asyncBufferProvider) Size() (int, error) {
	return d.b.Wait()
}

// Error returns any error that occurred during reading data from
// async buffer or the underlying source.
func (d *asyncBufferProvider) Error() error {
	if err := d.b.Error(); err != nil {
		return wrapDownloadError(err, d.desc)
	}
	return nil
}

// Close closes the async buffer and releases any resources held by it
func (d *asyncBufferProvider) Close() error {
	return d.b.Close()
}
