package fs

import (
	"io"
	"net/http"
)

type fileLimiter struct {
	f    http.File
	left int
}

func (lr *fileLimiter) Read(p []byte) (n int, err error) {
	if lr.left <= 0 {
		return 0, io.EOF
	}
	if len(p) > lr.left {
		p = p[0:lr.left]
	}
	n, err = lr.f.Read(p)
	lr.left -= n
	return
}

func (lr *fileLimiter) Close() error {
	return lr.f.Close()
}
