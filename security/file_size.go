package security

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var ErrSourceFileTooBig = ierrors.New(422, "Source image file is too big", "Invalid source image")

type hardLimitReader struct {
	r    io.Reader
	left int
}

func (lr *hardLimitReader) Read(p []byte) (n int, err error) {
	if lr.left <= 0 {
		return 0, ErrSourceFileTooBig
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
		return ErrSourceFileTooBig
	}

	return nil
}

func LimitFileSize(r io.Reader, opts Options) io.Reader {
	if opts.MaxSrcFileSize > 0 {
		return &hardLimitReader{r: r, left: opts.MaxSrcFileSize}
	}

	return r
}
