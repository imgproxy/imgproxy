package imagedata

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/imgproxy/imgproxy/v3/asyncbuffer"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	Watermark ImageData
)

// Provider represents the data of an image that can be read from a source.
// Please note that this interface can be backed by any reader, including lazy AsyncBuffer.
// There is no other way to guarantee that the data is read without errors except reading it till EOF.
type Provider interface {
	io.Closer

	Reader() io.ReadSeeker // Reader returns a new ReadSeeker for the image data
	Size() (int, error)    // Size returns the size of the image data in bytes
	Error() error          // Error returns any error that occurred during reading data from source
}

// ImageData is a provider with refcounting
type ImageData interface {
	Provider  // Provider provides access to the image data and metadata
	io.Closer // Must be closeable

	Format() imagetype.Type // Format returns the image format from the metadata (shortcut)
	Ref() ImageData         // Ref returns a new reference to the same image data

	// AddCancel attaches a cancel function to the image data.
	// Please note that Cancel functions must be idempotent: for instance, an implementation
	// could wrap cancel into sync.Once.
	AddCancel(cancel context.CancelFunc)
}

// imageData is a struct that implements the ImageData interface in full.
type imageData struct {
	Provider

	format   imagetype.Type
	mu       sync.Mutex
	cancel   []context.CancelFunc
	refCount atomic.Int32
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

// newImageData creates a new ImageData instance with the provided provider and format
func newImageData(provider Provider, format imagetype.Type) *imageData {
	d := &imageData{
		Provider: provider,
		format:   format,
		cancel:   nil,
	}

	d.refCount.Store(1)
	runtime.SetFinalizer(d, (*imageData).finalize)

	return d
}

// Ref returns a new reference to the same image data. It increments the reference count.
func (d *imageData) Ref() ImageData {
	for {
		old := d.refCount.Load()
		if old <= 0 {
			panic("imageData: Ref() called on closed imageData")
		}
		if d.refCount.CompareAndSwap(old, old+1) {
			return d
		}
	}
}

// Format returns the image format based on the metadata
func (d *imageData) Format() imagetype.Type {
	if d.refCount.Load() <= 0 {
		panic("imageData: Format() called on closed imageData")
	}

	return d.format
}

// Reader returns an io.ReadSeeker for the image data, but checks refcount first
func (d *imageData) Reader() io.ReadSeeker {
	if d.refCount.Load() <= 0 {
		panic("imageData: Reader() called on closed imageData")
	}

	return d.Provider.Reader()
}

// Size returns the size of the image data in bytes, but checks refcount first
func (d *imageData) Size() (int, error) {
	if d.refCount.Load() <= 0 {
		panic("imageData: Size() called on closed imageData")
	}

	return d.Provider.Size()
}

// AddCancel attaches a cancel function to the image data
func (d *imageData) AddCancel(cancel context.CancelFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.cancel = append(d.cancel, cancel)
}

// Close closes the image data and releases any resources held by it
func (d *imageData) Close() error {
	newRefCount := d.refCount.Add(-1)

	if newRefCount < 0 {
		panic("imageData: Close() called on already closed imageData")
	}

	if newRefCount > 0 {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.Provider.Close()

	cancels := d.cancel
	d.cancel = nil

	for _, cancel := range cancels {
		cancel()
	}

	return nil
}

// finalize is called by the GC when the imageData is collected. If the refcount
// is still positive the object was leaked — log a warning and release resources.
func (d *imageData) finalize() {
	if d.refCount.Load() <= 0 {
		return
	}
	slog.Warn("imageData: collected by GC without being closed (resource leak)")
	d.Provider.Close() //nolint:errcheck
	for _, cancel := range d.cancel {
		cancel()
	}
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
