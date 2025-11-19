package storage

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Reader represents a generic storage interface, which can read
// objects from a storage backend.
type Reader interface {
	// GetObject retrieves an object from the storage and returns
	// ObjectReader with the result.
	GetObject(
		ctx context.Context,
		reqHeader http.Header,
		bucket, key, query string,
	) (*ObjectReader, error)
}

// ObjectReader represents a generic reader for a storage object.
// It can be in any state: success, error, not found, etc.
// It can be converted to HTTP response or used as-is.
type ObjectReader struct {
	Status        int           // HTTP status code
	Headers       http.Header   // Response headers harvested from the engine response
	Body          io.ReadCloser // Response body reader
	contentLength int64
}

// NewObjectOK creates a new Reader with a 200 OK status.
func NewObjectOK(headers http.Header, body io.ReadCloser) *ObjectReader {
	return &ObjectReader{
		Status:        http.StatusOK,
		Headers:       headers,
		Body:          body,
		contentLength: -1, // is set in Response()
	}
}

// NewObjectPartialContent creates a new Reader with a 206 Partial Content status.
func NewObjectPartialContent(headers http.Header, body io.ReadCloser) *ObjectReader {
	return &ObjectReader{
		Status:        http.StatusPartialContent,
		Headers:       headers,
		Body:          body,
		contentLength: -1, // is set in Response()
	}
}

// NewObjectNotFound creates a new Reader with a 404 Not Found status.
func NewObjectNotFound(message string) *ObjectReader {
	return NewObjectError(http.StatusNotFound, message)
}

// NewObjectError creates a new Reader with a custom status code
func NewObjectError(statusCode int, message string) *ObjectReader {
	return &ObjectReader{
		Status:        statusCode,
		Body:          io.NopCloser(strings.NewReader(message)),
		Headers:       http.Header{httpheaders.ContentType: {"text/plain"}},
		contentLength: int64(len(message)),
	}
}

// NewObjectNotModified creates a new Reader with a 304 Not Modified status.
func NewObjectNotModified(headers http.Header) *ObjectReader {
	// Copy headers relevant to NotModified response only
	nmHeaders := make(http.Header)
	httpheaders.Copy(
		headers,
		nmHeaders,
		[]string{httpheaders.Etag, httpheaders.LastModified},
	)

	return &ObjectReader{
		Status:        http.StatusNotModified,
		Headers:       nmHeaders,
		contentLength: 0,
	}
}

// NewInvalidRang creates a new Reader with a 416 Range Not Satisfiable status.
func NewObjectInvalidRange() *ObjectReader {
	return &ObjectReader{
		Status:        http.StatusRequestedRangeNotSatisfiable,
		contentLength: 0,
	}
}

// ContentLength returns the content length of the response.
func (r *ObjectReader) ContentLength() int64 {
	if r.contentLength > 0 {
		return r.contentLength
	}

	h := r.Headers.Get(httpheaders.ContentLength)
	if len(h) > 0 {
		p, err := strconv.ParseInt(h, 10, 64)
		if err != nil {
			return p
		}
	}

	return -1
}

// Response converts Reader to http.Response
func (r *ObjectReader) Response(req *http.Request) *http.Response {
	return &http.Response{
		Status:        http.StatusText(r.Status),
		StatusCode:    r.Status,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        r.Headers,
		Body:          r.Body,
		Close:         true,
		Request:       req,
		ContentLength: r.ContentLength(),
	}
}
