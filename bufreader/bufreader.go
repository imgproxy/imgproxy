// bufreader provides a buffered reader that reads from io.Reader, but caches
// the data in a bytes.Buffer to allow peeking and discarding without re-reading.
package bufreader

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/ioutil"
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

	finished bool // Indicates if the reader has reached EOF
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

	if br.pos >= len(br.buf) {
		return 0, io.EOF // No more data to read
	}

	n := copy(p, br.buf[br.pos:])
	br.pos += n
	return n, nil
}

// Peek returns the next n bytes from the buffered reader without advancing the position.
func (br *Reader) Peek(n int) ([]byte, error) {
	if err := br.fetch(br.pos + n); err != nil {
		return nil, err
	}

	if br.pos >= len(br.buf) {
		return nil, io.EOF // No more data to read
	}

	// Return slice of buffered data without advancing position
	available := br.buf[br.pos:]
	return available[:min(len(available), n)], nil
}

// Rewind seeks buffer to the beginning
func (br *Reader) Rewind() {
	br.pos = 0
}

// fetch ensures the buffer contains at least 'need' bytes
func (br *Reader) fetch(need int) error {
	if br.finished || need <= len(br.buf) {
		return nil
	}

	b := make([]byte, need-len(br.buf))
	n, err := ioutil.TryReadFull(br.r, b)

	if err == io.EOF {
		// If we reached EOF, we mark the reader as finished
		br.finished = true
	} else if err != nil {
		return err
	}

	if n > 0 {
		// append only those which we read in fact
		br.buf = append(br.buf, b[:n]...)
	}

	return nil
}
