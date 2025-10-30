package response

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Object represents a generic response for a storage object.
// It can be converted to HTTP response or used as-is.
type Object struct {
	Status        int           // HTTP status code
	Headers       http.Header   // Response headers harvested from the engine response
	Body          io.ReadCloser // Response body reader
	contentLength int64
}

// NewOK creates a new ObjectReader with a 200 OK status.
func NewOK(headers http.Header, body io.ReadCloser) *Object {
	return &Object{
		Status:        http.StatusOK,
		Headers:       headers,
		Body:          body,
		contentLength: -1, // is set in Response()
	}
}

// NewPartialContent creates a new ObjectReader with a 206 Partial Content status.
func NewPartialContent(headers http.Header, body io.ReadCloser) *Object {
	return &Object{
		Status:        http.StatusPartialContent,
		Headers:       headers,
		Body:          body,
		contentLength: -1, // is set in Response()
	}
}

// NewNotFound creates a new ObjectReader with a 404 Not Found status.
func NewNotFound(message string) *Object {
	return NewError(http.StatusNotFound, message)
}

// NewError creates a new ObjectReader with a custom status code
func NewError(statusCode int, message string) *Object {
	return &Object{
		Status:        statusCode,
		Body:          io.NopCloser(strings.NewReader(message)),
		Headers:       http.Header{httpheaders.ContentType: {"text/plain"}},
		contentLength: int64(len(message)),
	}
}

// NewNotModified creates a new ObjectReader with a 304 Not Modified status.
func NewNotModified(headers http.Header) *Object {
	// Copy headers relevant to NotModified response only
	nmHeaders := make(http.Header)
	httpheaders.Copy(
		headers,
		nmHeaders,
		[]string{httpheaders.Etag, httpheaders.LastModified},
	)

	return &Object{
		Status:        http.StatusNotModified,
		Headers:       nmHeaders,
		contentLength: 0,
	}
}

// NewInvalidRang creates a new ObjectReader with a 416 Range Not Satisfiable status.
func NewInvalidRange() *Object {
	return &Object{
		Status:        http.StatusRequestedRangeNotSatisfiable,
		contentLength: 0,
	}
}

// ContentLength returns the content length of the response.
func (r *Object) ContentLength() int64 {
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

// Response converts ObjectReader to http.Response
func (r *Object) Response(req *http.Request) *http.Response {
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
