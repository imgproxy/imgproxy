package security

import (
	"io"
	"net/http"
	"os"
)

// hardLimitReadCloser is a wrapper around io.ReadCloser
// that limits the number of bytes it can read from the upstream reader.
type hardLimitReadCloser struct {
	r    io.ReadCloser
	left int
}

// Read reads data from the underlying reader, limiting the number of bytes read
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

// Close closes the underlying reader
func (lr *hardLimitReadCloser) Close() error {
	return lr.r.Close()
}

// LimitFileSize limits the size of the file to MaxSrcFileSize (if set).
// It calls f.Stat() to get the file to get its size and returns an error
// if the size exceeds MaxSrcFileSize.
func LimitFileSize(f *os.File, opts Options) (*os.File, error) {
	if opts.MaxSrcFileSize == 0 {
		return f, nil
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if int(s.Size()) > opts.MaxSrcFileSize {
		return nil, newFileSizeError()
	}

	return f, nil
}

// LimitResponseSize limits the size of the response body to MaxSrcFileSize (if set).
// First, it tries to use Content-Length header to check the limit.
// If Content-Length is not set, it limits the size of the response body by wrapping
// body reader with hard limit reader.
func LimitResponseSize(r *http.Response, opts Options) (*http.Response, error) {
	if opts.MaxSrcFileSize == 0 {
		return r, nil
	}

	// If Content-Length was set, limit the size of the response body before reading it
	size := int(r.ContentLength)

	if size > opts.MaxSrcFileSize {
		return nil, newFileSizeError()
	}

	// hard-limit the response body reader
	r.Body = &hardLimitReadCloser{r: r.Body, left: opts.MaxSrcFileSize}

	return r, nil
}
