package security

import (
	"io"
	"net/http"
)

// hardLimitReadCloser is a wrapper around io.ReadCloser
// that limits the number of bytes it can read from the upstream reader.
type hardLimitReadCloser struct {
	r    io.ReadCloser
	left int
}

func (lr *hardLimitReadCloser) Read(p []byte) (n int, err error) {
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

func (lr *hardLimitReadCloser) Close() error {
	return lr.r.Close()
}

// LimitResponseSize limits the size of the response body to MaxSrcFileSize (if set).
// First, it tries to use Content-Length header to check the limit.
// If Content-Length is not set, it limits the size of the response body by wrapping
// body reader with hard limit reader.
func LimitResponseSize(r *http.Response, limit int) (*http.Response, error) {
	if limit == 0 {
		return r, nil
	}

	// If Content-Length was set, limit the size of the response body before reading it
	size := int(r.ContentLength)
	if size > limit {
		return nil, newFileSizeError()
	}

	// hard-limit the response body reader
	r.Body = &hardLimitReadCloser{r: r.Body, left: limit}

	return r, nil
}
