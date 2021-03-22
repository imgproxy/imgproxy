package main

import (
	"bufio"
	"bytes"
	"io"
)

type bufReader struct {
	r   io.Reader
	buf *bytes.Buffer
	cur int
}

func newBufReader(r io.Reader, buf *bytes.Buffer) *bufReader {
	br := bufReader{
		r:   r,
		buf: buf,
	}
	return &br
}

func (br *bufReader) Read(p []byte) (int, error) {
	if err := br.fill(br.cur + len(p)); err != nil {
		return 0, err
	}

	n := copy(p, br.buf.Bytes()[br.cur:])
	br.cur += n
	return n, nil
}

func (br *bufReader) ReadByte() (byte, error) {
	if err := br.fill(br.cur + 1); err != nil {
		return 0, err
	}

	b := br.buf.Bytes()[br.cur]
	br.cur++
	return b, nil
}

func (br *bufReader) Discard(n int) (int, error) {
	if n < 0 {
		return 0, bufio.ErrNegativeCount
	}
	if n == 0 {
		return 0, nil
	}

	if err := br.fill(br.cur + n); err != nil {
		return 0, err
	}

	n = minInt(n, br.buf.Len()-br.cur)
	br.cur += n
	return n, nil
}

func (br *bufReader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return []byte{}, bufio.ErrNegativeCount
	}
	if n == 0 {
		return []byte{}, nil
	}

	if err := br.fill(br.cur + n); err != nil {
		return []byte{}, err
	}

	if n > br.buf.Len()-br.cur {
		return br.buf.Bytes()[br.cur:], io.EOF
	}

	return br.buf.Bytes()[br.cur : br.cur+n], nil
}

func (br *bufReader) Flush() error {
	_, err := br.buf.ReadFrom(br.r)
	return err
}

func (br *bufReader) fill(need int) error {
	n := need - br.buf.Len()
	if n <= 0 {
		return nil
	}

	n = maxInt(4096, n)

	if _, err := br.buf.ReadFrom(io.LimitReader(br.r, int64(n))); err != nil {
		return err
	}

	// Nothing was read, it's EOF
	if br.cur == br.buf.Len() {
		return io.EOF
	}

	return nil
}
