package imagedata

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	Watermark ImageData

	panicOnLeak bool
)

func init() {
	panicOnLeak = len(os.Getenv("TEST_IMAGEDATA_REFCOUNT_PANIC")) > 0
}

// ImageData is a provider with refcounting
type ImageData interface {
	Provider // Provider provides access to the image data and metadata

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
		old := d.checkRef("Ref")
		if d.refCount.CompareAndSwap(old, old+1) {
			return d
		}
	}
}

// Format returns the image format based on the metadata
func (d *imageData) Format() imagetype.Type {
	d.checkRef("Format")

	return d.format
}

// Reader returns an io.ReadSeeker for the image data, but checks refcount first
func (d *imageData) Reader() io.ReadSeeker {
	d.checkRef("Reader")

	return d.Provider.Reader()
}

// Size returns the size of the image data in bytes, but checks refcount first
func (d *imageData) Size() (int, error) {
	d.checkRef("Size")

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

// checkRef panics if the imageData has already been closed and returns the current refcount.
func (d *imageData) checkRef(method string) int32 {
	v := d.refCount.Load()
	if v <= 0 {
		panic(fmt.Sprintf("imageData: %s() called on closed imageData", method))
	}
	return v
}

// finalize is called by the GC when the imageData is collected. If the refcount
// is still positive the object was leaked — log a warning and release resources.
func (d *imageData) finalize() {
	if d.refCount.Load() <= 0 {
		return
	}
	if panicOnLeak {
		panic("imageData: collected by GC without being closed (resource leak)")
	}
	slog.Warn("imageData: collected by GC without being closed (resource leak)")
	d.Provider.Close() //nolint:errcheck
	for _, cancel := range d.cancel {
		cancel()
	}
}
