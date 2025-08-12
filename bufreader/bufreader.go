// bufreader provides a buffered reader that reads from io.Reader, but caches
// the data in a bytes.Buffer to allow peeking and discarding without re-reading.
package bufreader

import (
	"io"
)

// ReadPeeker is an interface that combines io.Reader and a method to peek at the next n bytes
type ReadPeeker interface {
	io.Reader
	Peek(n int) ([]byte, error) // Peek returns the next n bytes without advancing
}

// Reader is a buffered reader that reads from an io.Reader and caches the data.
type Reader struct {
	r   io.Reader
	buf []byte
	pos int
}

// New creates new buffered reader
func New(r io.Reader) *Reader {
	br := Reader{
		r:   r,
		buf: nil,
	}
	return &br
}

// Read reads data into p from the buffered reader.
func (br *Reader) Read(p []byte) (int, error) {
	if err := br.fetch(br.pos + len(p)); err != nil {
		return 0, err
	}

	n := copy(p, br.buf[br.pos:])
	br.pos += n
	return n, nil
}

// Peek returns the next n bytes from the buffered reader without advancing the position.
func (br *Reader) Peek(n int) ([]byte, error) {
	err := br.fetch(br.pos + n)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Return slice of buffered data without advancing position
	available := br.buf[br.pos:]
	if len(available) == 0 && err == io.EOF {
		return nil, io.EOF
	}

	return available[:min(len(available), n)], nil
}

// Rewind seeks buffer to the beginning
func (br *Reader) Rewind() {
	br.pos = 0
}

// fetch ensures the buffer contains at least 'need' bytes
func (br *Reader) fetch(need int) error {
	if need-len(br.buf) <= 0 {
		return nil
	}

	b := make([]byte, need)
	n, err := io.ReadFull(br.r, b)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}

	// append only those which we read in fact
	br.buf = append(br.buf, b[:n]...)

	return nil
}
