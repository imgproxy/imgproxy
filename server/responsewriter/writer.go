package responsewriter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Just aliases for [http.ResponseWriter] and [http.ResponseController].
// We need them to make them private in [Writer] so they can't be accessed directly.
type httpResponseWriter = http.ResponseWriter
type httpResponseController = *http.ResponseController

// Writer is an implementation of [http.ResponseWriter] with additional
// functionality for managing response headers.
type Writer struct {
	httpResponseWriter
	httpResponseController

	config        *Config     // Configuration for the writer
	originHeaders http.Header // Original response headers
	result        http.Header // Headers to be written to the response
	maxAge        int         // Current max age for Cache-Control header

	beforeWriteOnce sync.Once
}

// HTTPResponseWriter returns the underlying http.ResponseWriter.
func (w *Writer) HTTPResponseWriter() http.ResponseWriter {
	return w.httpResponseWriter
}

// SetHTTPResponseWriter replaces the underlying http.ResponseWriter.
func (w *Writer) SetHTTPResponseWriter(rw http.ResponseWriter) {
	w.httpResponseWriter = rw
	w.httpResponseController = http.NewResponseController(rw)
}

// SetOriginHeaders sets the origin headers for the request.
func (w *Writer) SetOriginHeaders(h http.Header) {
	w.originHeaders = h
}

// SetIsFallbackImage sets the Fallback-Image header to
// indicate that the fallback image was used.
func (w *Writer) SetIsFallbackImage() {
	// We set maxAge to FallbackImageTTL if it's explicitly passed
	if w.config.FallbackImageTTL < 0 {
		return
	}

	// However, we should not overwrite existing value if set (or greater than ours)
	if w.maxAge < 0 || w.maxAge > w.config.FallbackImageTTL {
		w.maxAge = w.config.FallbackImageTTL
	}
}

// SetExpires sets the TTL from time
func (w *Writer) SetExpires(expires time.Time) {
	if expires.IsZero() {
		return
	}

	// Convert current maxAge to time
	currentMaxAgeTime := time.Now().Add(time.Duration(w.maxAge) * time.Second)

	// If maxAge outlives expires or was not set, we'll use expires as maxAge.
	if w.maxAge < 0 || expires.Before(currentMaxAgeTime) {
		w.maxAge = min(w.config.DefaultTTL, max(0, int(time.Until(expires).Seconds())))
	}
}

// SetVary sets the Vary header
func (w *Writer) SetVary() {
	if val := w.config.VaryValue; len(val) > 0 {
		w.result.Set(httpheaders.Vary, val)
	}
}

// SetContentDisposition sets the Content-Disposition header, passthrough to ContentDispositionValue
func (w *Writer) SetContentDisposition(originURL, filename, ext, contentType string, returnAttachment bool) {
	value := httpheaders.ContentDispositionValue(
		originURL,
		filename,
		ext,
		contentType,
		returnAttachment,
	)

	if value != "" {
		w.result.Set(httpheaders.ContentDisposition, value)
	}
}

// Passthrough copies specified headers from the original response headers to the response headers.
func (w *Writer) Passthrough(only ...string) {
	httpheaders.Copy(w.originHeaders, w.result, only)
}

// CopyFrom copies specified headers from the headers object. Please note that
// all the past operations may overwrite those values.
func (w *Writer) CopyFrom(headers http.Header, only []string) {
	httpheaders.Copy(headers, w.result, only)
}

// SetContentLength sets the Content-Length header
func (w *Writer) SetContentLength(contentLength int) {
	if contentLength < 0 {
		return
	}

	w.result.Set(httpheaders.ContentLength, strconv.Itoa(contentLength))
}

// SetContentType sets the Content-Type header
func (w *Writer) SetContentType(mime string) {
	w.result.Set(httpheaders.ContentType, mime)
}

// writeCanonical sets the Link header with the canonical URL.
// It is mandatory for any response if enabled in the configuration.
func (w *Writer) SetCanonical(url string) {
	if !w.config.SetCanonicalHeader {
		return
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		value := fmt.Sprintf(`<%s>; rel="canonical"`, url)
		w.result.Set(httpheaders.Link, value)
	}
}

// setCacheControl sets the Cache-Control header with the specified value.
func (w *Writer) setCacheControl(value int) bool {
	if value <= 0 {
		return false
	}

	w.result.Set(httpheaders.CacheControl, fmt.Sprintf("max-age=%d, public", value))
	return true
}

// setCacheControlNoCache sets the Cache-Control header to no-cache (default).
func (w *Writer) setCacheControlNoCache() {
	w.result.Set(httpheaders.CacheControl, "no-cache")
}

// setCacheControlPassthrough sets the Cache-Control header from the request
// if passthrough is enabled in the configuration.
func (w *Writer) setCacheControlPassthrough() bool {
	if !w.config.CacheControlPassthrough || w.maxAge > 0 {
		return false
	}

	if val := w.originHeaders.Get(httpheaders.CacheControl); val != "" {
		w.result.Set(httpheaders.CacheControl, val)
		return true
	}

	if val := w.originHeaders.Get(httpheaders.Expires); val != "" {
		if t, err := time.Parse(http.TimeFormat, val); err == nil {
			maxAge := max(0, int(time.Until(t).Seconds()))
			return w.setCacheControl(maxAge)
		}
	}

	return false
}

// setCSP sets the Content-Security-Policy header to prevent script execution.
func (w *Writer) setCSP() {
	w.result.Set(httpheaders.ContentSecurityPolicy, "script-src 'none'")
}

// flushHeaders writes the headers to the response writer. It does not overwrite
// target headers, which were set outside the header writer.
func (w *Writer) flushHeaders() {
	// Then, let's try to set Cache-Control using priority order
	switch {
	case w.setCacheControl(w.maxAge): // First, try set explicit
	case w.setCacheControlPassthrough(): // Try to pick up from request headers
	case w.setCacheControl(w.config.DefaultTTL): // Fallback to default value
	default:
		w.setCacheControlNoCache() // By default we use no-cache
	}

	w.setCSP()

	// Copy all headers to the response without overwriting existing ones
	httpheaders.CopyAll(w.result, w.Header(), false)
}

// beforeWrite is called before [WriteHeader] and [Write]
func (w *Writer) beforeWrite() {
	w.beforeWriteOnce.Do(func() {
		// We're going to start writing response.
		// Set write deadline.
		w.SetWriteDeadline(time.Now().Add(w.config.WriteResponseTimeout))

		// Flush headers before we write anything
		w.flushHeaders()
	})
}

// WriteHeader writes the HTTP response header.
//
// It ensures that all headers are flushed before writing the status code.
func (w *Writer) WriteHeader(statusCode int) {
	w.beforeWrite()

	w.httpResponseWriter.WriteHeader(statusCode)
}

// Write writes the HTTP response body.
//
// It ensures that all headers are flushed before writing the body.
func (w *Writer) Write(b []byte) (int, error) {
	w.beforeWrite()

	return w.httpResponseWriter.Write(b)
}
