package asyncbuffer

import (
	"errors"
	"io"
)

// Underlying Reader that provides io.ReadSeeker interface for the actual data reading
// What is the purpose of this Reader?
type Reader struct {
	ab  *AsyncBuffer
	pos int64
}

// Read reads data from the AsyncBuffer.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.ab.readAt(p, r.pos)
	if err == nil {
		r.pos += int64(n)
	}

	return n, err
}

// Seek sets the position of the reader to the given offset and returns the new position
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.pos = offset

	case io.SeekCurrent:
		r.pos += offset

	case io.SeekEnd:
		size, err := r.ab.Wait()
		if err != nil {
			return 0, err
		}

		r.pos = int64(size) + offset

	default:
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: invalid whence")
	}

	if r.pos < 0 {
		return 0, errors.New("asyncbuffer.AsyncBuffer.ReadAt: negative position")
	}

	return r.pos, nil
}
