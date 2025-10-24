package xmlparser

import (
	"bytes"
	"math"
	"slices"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		// Reserve some capacity to not re-allocate on short strings.
		buf := make(buffer, 0, 1024)
		return &buf
	},
}

// buffer is a slice of bytes with some additional convenience methods.
type buffer []byte

// newBuffer creates a new buffer from the pool.
func newBuffer() *buffer {
	return bufPool.Get().(*buffer)
}

// Free truncates the buffer and returns it to the pool.
func (b *buffer) Free() {
	// Don't keep large buffers around.
	if len(*b) > 16*1024 {
		return
	}

	b.Reset()
	bufPool.Put(b)
}

// Reset truncates the buffer.
func (b *buffer) Reset() {
	*b = (*b)[:0]
}

func (b *buffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.grow(len(p))
	*b = append(*b, p...)
	return len(p), nil
}

// WriteByte writes a single byte to the buffer.
func (b *buffer) WriteByte(p byte) error {
	b.grow(1)
	*b = append(*b, p)
	return nil
}

// grow ensures that the buffer has at least n free bytes of capacity.
// It grows the buffer to the next power of two if necessary.
func (b *buffer) grow(n int) {
	if req := len(*b) + n; req > cap(*b) {
		p := math.Ceil(math.Log2(float64(req)))
		*b = slices.Grow(*b, 1<<int(p))
	}
}

// HasSuffix reports whether the buffer ends with the given suffix.
func (b *buffer) HasSuffix(suffix []byte) bool {
	return bytes.HasSuffix(*b, suffix)
}

// Remove removes the last n bytes.
func (b *buffer) Remove(n int) {
	l := len(*b)
	n = min(n, l)
	*b = (*b)[:l-n]
}

// Bytes returns the contents of the buffer as a byte slice.
func (b *buffer) Bytes() []byte {
	return *b
}

// String returns the contents of the buffer as a string.
func (b *buffer) String() string {
	return string(*b)
}

// len returns the number of bytes written to the buffer.
func (b *buffer) Len() int {
	return len(*b)
}
