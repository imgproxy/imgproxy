package imagedata

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/imgproxy/imgproxy/v4/asyncbuffer"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/imath"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// NewFromReaderAsync creates an ImageData that reads from r asynchronously and
// lazily, backed by an AsyncBuffer: the caller only blocks until enough data
// is available to detect the image format, the rest is read in the
// background as it's consumed. dataLen is the expected length of the data in
// r (<=0 if unknown). r is closed once the buffer finishes reading. Each
// finishFn is called exactly once when the buffer finishes.
func NewFromReaderAsync(
	r io.ReadCloser, dataLen int, ct, ext, desc string, finishFn ...context.CancelFunc,
) (ImageData, error) {
	b := asyncbuffer.New(r, dataLen, finishFn...)

	format, err := imagetype.Detect(b.Reader(), ct, ext)
	if err != nil {
		b.Close()
		return nil, err
	}

	// We successfully detected the image type, so we can release the pause
	// and let the buffer read the rest of the data immediately.
	b.ReleaseThreshold()

	return newImageData(&asyncBufferProvider{b: b, desc: desc}, format), nil
}

// NewFromReaderSync creates an ImageData by eagerly reading all of r into an
// in-memory buffer (pooled for reuse), detecting the image format along the
// way via a TeeReader. dataLen is the expected length of the data in r
// (<=0 if unknown); maxSize additionally caps how much is preallocated
// (<=0 for no cap). r is read to completion; the caller remains responsible
// for closing r.
func NewFromReaderSync(r io.Reader, dataLen, maxSize int, ct, ext string) (ImageData, error) {
	buf := bufPool.Get().(*bytes.Buffer) //nolint:forcetypeassert
	buf.Reset()

	cancel := func() {
		bufPool.Put(buf)
	}

	// Create a TeeReader to write to buffer while reading.
	tr := io.TeeReader(r, buf)

	// Detect image type using the TeeReader.
	format, err := imagetype.Detect(tr, ct, ext)
	if err != nil {
		cancel()
		return nil, err
	}

	// Preallocate buffer size if the data length is known.
	growLen := imath.MinNonZero(dataLen, maxSize) - buf.Len()
	if growLen > 0 {
		buf.Grow(growLen)
	}

	// Read the rest of the data into the buffer
	if _, err := buf.ReadFrom(r); err != nil {
		cancel()
		return nil, err
	}

	// Create ImageData from the buffer bytes and add the cancel function
	// to return the buffer to the pool when done.
	d := NewFromBytesWithFormat(format, buf.Bytes())
	d.AddCancel(cancel)

	return d, nil
}
