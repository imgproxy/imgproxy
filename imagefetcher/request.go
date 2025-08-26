package imagefetcher

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

var (
	// contentRangeRe Content-Range header regex to check if the response is a partial content response
	contentRangeRe = regexp.MustCompile(`^bytes ((\d+)-(\d+)|\*)/(\d+|\*)$`)
)

// Request is a struct that holds the request and cancel function for an image fetcher request
type Request struct {
	fetcher *Fetcher           // Parent ImageFetcher instance
	request *http.Request      // HTTP request to fetch the image
	cancel  context.CancelFunc // Request context cancel function
}

// Send sends the generic request and returns the http.Response or an error
func (r *Request) Send() (*http.Response, error) {
	client := r.fetcher.newHttpClient()

	// Let's add a cookie jar to the client if the request URL is HTTP or HTTPS
	// This is necessary to pass cookie challenge for some servers.
	if r.request.URL.Scheme == "http" || r.request.URL.Scheme == "https" {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		client.Jar = jar
	}

	for {
		// Try request
		res, err := client.Do(r.request)
		if err == nil {
			return res, nil // Return successful response
		}

		// Close the response body if request was unsuccessful
		if res != nil && res.Body != nil {
			res.Body.Close()
		}

		// Retry if the error is due to a lost connection
		if strings.Contains(err.Error(), connectionLostError) {
			select {
			case <-r.request.Context().Done():
				return nil, err
			case <-time.After(bounceDelay):
				continue
			}
		}

		return nil, WrapError(err)
	}
}

// FetchImage fetches the image using the request and returns the response or an error.
// It checks for the NotModified status and handles partial content responses.
func (r *Request) FetchImage() (*http.Response, error) {
	res, err := r.Send()
	if err != nil {
		return nil, err
	}

	// If the source image was not modified, close the body and NotModifiedError
	if res.StatusCode == http.StatusNotModified {
		res.Body.Close()
		return nil, newNotModifiedError(res.Header)
	}

	// If the source responds with 206, check if the response contains an entire image.
	// If not, return an error.
	if res.StatusCode == http.StatusPartialContent {
		err = checkPartialContentResponse(res)
		if err != nil {
			res.Body.Close()
			return nil, err
		}
	} else if res.StatusCode != http.StatusOK {
		body := extractErraticBody(res)
		res.Body.Close()
		return nil, newImageResponseStatusError(res.StatusCode, body)
	}

	// If the response is gzip encoded, wrap it in a gzip reader
	err = wrapGzipBody(res)
	if err != nil {
		res.Body.Close()
		return nil, err
	}

	return res, nil
}

// Cancel cancels the request context
func (r *Request) Cancel() {
	r.cancel()
}

// URL returns the actual URL of the request
func (r *Request) URL() *url.URL {
	return r.request.URL
}

// checkPartialContentResponse if the response is a partial content response,
// we check if it contains the entire image.
func checkPartialContentResponse(res *http.Response) error {
	contentRange := res.Header.Get(httpheaders.ContentRange)
	rangeParts := contentRangeRe.FindStringSubmatch(contentRange)

	if len(rangeParts) == 0 {
		return newImagePartialResponseError("Partial response with invalid Content-Range header")
	}

	if rangeParts[1] == "*" || rangeParts[2] != "0" {
		return newImagePartialResponseError("Partial response with incomplete content")
	}

	contentLengthStr := rangeParts[4]
	if contentLengthStr == "*" {
		contentLengthStr = res.Header.Get(httpheaders.ContentLength)
	}

	contentLength, _ := strconv.Atoi(contentLengthStr)
	rangeEnd, _ := strconv.Atoi(rangeParts[3])

	if contentLength <= 0 || rangeEnd != contentLength-1 {
		return newImagePartialResponseError("Partial response with incomplete content")
	}

	return nil
}

// extractErraticBody extracts the error body from the response if it is a text-based content type
func extractErraticBody(res *http.Response) string {
	if strings.HasPrefix(res.Header.Get(httpheaders.ContentType), "text/") {
		bbody, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return string(bbody)
	}

	return ""
}

// wrapGzipBody wraps the response body in a gzip reader if the Content-Encoding is gzip.
// We set DisableCompression: true to avoid sending the Accept-Encoding: gzip header,
// since we do not want to compress image data (which is usually already compressed).
// However, some servers still send gzip-encoded responses regardless.
func wrapGzipBody(res *http.Response) error {
	if res.Header.Get(httpheaders.ContentEncoding) == "gzip" {
		gzipBody, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil
		}
		res.Body = &gzipReadCloser{
			Reader: gzipBody,
			r:      res.Body,
		}
		res.Header.Del(httpheaders.ContentEncoding)
	}

	return nil
}

// gzipReadCloser is a wrapper around gzip.Reader which also closes the original body
type gzipReadCloser struct {
	*gzip.Reader
	r io.ReadCloser
}

// Close closes the gzip reader and the original body
func (gr *gzipReadCloser) Close() error {
	gr.Reader.Close()
	return gr.r.Close()
}
