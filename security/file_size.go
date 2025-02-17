package security

import (
	"io"
)

type hardLimitReader struct {
	r    io.Reader
	left int
}

func (lr *hardLimitReader) Read(p []byte) (n int, err error) {
	if lr.left <= 0 {
		return 0, newFileSizeError()
	}
	if len(p) > lr.left {
		p = p[0:lr.left]
	}
	n, err = lr.r.Read(p)
	lr.left -= n
	return
}

func CheckFileSize(size int, opts Options) error {
	if opts.MaxSrcFileSize > 0 && size > opts.MaxSrcFileSize {
		return newFileSizeError()
	}

	return nil
}

func LimitFileSize(r io.Reader, opts Options) io.Reader {
	if opts.MaxSrcFileSize > 0 {
		return &hardLimitReader{r: r, left: opts.MaxSrcFileSize}
	}

	return r
}
