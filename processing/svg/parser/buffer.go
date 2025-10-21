package svgparser

import (
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

// WriteByte writes a single byte to the buffer.
func (b *buffer) WriteByte(p byte) error {
	if c := cap(*b); len(*b) == c {
		*b = slices.Grow(*b, c)
	}
	*b = append(*b, p)
	return nil
}

// Last returns the last n bytes of the buffer.
// If n is greater than the buffer length, the entire buffer is returned.
func (b *buffer) Last(n int) []byte {
	l := len(*b)
	n = min(n, l)
	return (*b)[l-n:]
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
