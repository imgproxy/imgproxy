// headerwriter is responsible for writing processing/stream response headers
package headerwriter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Writer is a struct that builds HTTP response headers.
type Writer struct {
	config                  *Config
	originalResponseHeaders http.Header // Original response headers
	result                  http.Header // Headers to be written to the response
	maxAge                  int         // Current max age for Cache-Control header
	url                     string      // URL of the request, used for canonical header
	varyValue               string      // Vary header value
}

// New creates a new HeaderBuilder instance with the provided origin headers and URL
func New(config *Config, originalResponseHeaders http.Header, url string) *Writer {
	vary := make([]string, 0)

	if config.SetVaryAccept {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	varyValue := strings.Join(vary, ", ")

	return &Writer{
		config:                  config,
		originalResponseHeaders: originalResponseHeaders,
		url:                     url,
		result:                  make(http.Header),
		maxAge:                  -1,
		varyValue:               varyValue,
	}
}

// TODO: Do not remove, will be used shortly in processing_handler.go
//
// SetIsFallbackImage sets the Fallback-Image header to
// indicate that the fallback image was used.
// func (w *Writer) SetIsFallbackImage() {
// 	w.result.Set("Fallback-Image", "1")
// }

// SetMaxAge sets the max-age for the Cache-Control header.
//
// It accepts two values:
// - force: usually comes from ProcessingOptions.
// - ttl which is the time-to-live value.
//
// force is used if ttl is blank. ttl can't outlive force.
func (w *Writer) SetMaxAge(force *time.Time, ttl int) {
	if ttl > 0 {
		w.maxAge = ttl
	}

	if force == nil {
		return
	}

	// Convert current maxAge to time
	currentMaxAgeTime := time.Now().Add(time.Duration(w.maxAge) * time.Second)

	// If maxAge outlives expires or was not set, we'll use expires as maxAge.
	if w.maxAge < 0 || force.Before(currentMaxAgeTime) {
		// Get the TTL from the expires time (must not be in the past)
		expiresTTL := max(0, int(time.Until(*force).Seconds()))

		if expiresTTL > 0 {
			w.maxAge = expiresTTL
		}
	}
}

// SetLastModified sets the Last-Modified header from request
func (w *Writer) SetLastModified() {
	if !w.config.LastModifiedEnabled {
		return
	}

	val := w.originalResponseHeaders.Get(httpheaders.LastModified)
	if len(val) == 0 {
		return
	}

	w.result.Set(httpheaders.LastModified, val)
}

// SetVary sets the Vary header
func (w *Writer) SetVary() {
	if len(w.varyValue) > 0 {
		w.result.Set(httpheaders.Vary, w.varyValue)
	}
}

// Passthrough copies specified headers from the original response headers to the response headers.
func (w *Writer) Passthrough(only []string) {
	for _, key := range only {
		values := w.originalResponseHeaders.Values(key)

		for _, value := range values {
			w.result.Add(key, value)
		}
	}
}

// CopyFrom copies specified headers from the headers object. Please note that
// all the past operations may overwrite those values.
func (w *Writer) CopyFrom(headers http.Header, only []string) {
	for _, key := range only {
		values := headers.Values(key)

		for _, value := range values {
			w.result.Add(key, value)
		}
	}
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
func (b *Writer) SetCanonical() {
	if !b.config.SetCanonicalHeader {
		return
	}

	if strings.HasPrefix(b.url, "https://") || strings.HasPrefix(b.url, "http://") {
		value := fmt.Sprintf(`<%s>; rel="canonical"`, b.url)
		b.result.Set(httpheaders.Link, value)
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

	if val := w.originalResponseHeaders.Get(httpheaders.CacheControl); val != "" {
		w.result.Set(httpheaders.CacheControl, val)
		return true
	}

	if val := w.originalResponseHeaders.Get(httpheaders.Expires); val != "" {
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

// Write writes the headers to the response writer. It does not overwrite
// target headers, which were set outside the header writer.
func (w *Writer) Write(rw http.ResponseWriter) {
	// Then, let's try to set Cache-Control using priority order
	switch {
	case w.setCacheControl(w.maxAge): // First, try set explicit
	case w.setCacheControlPassthrough(): // Try to pick up from request headers
	case w.setCacheControl(w.config.DefaultTTL): // Fallback to default value
	default:
		w.setCacheControlNoCache() // By default we use no-cache
	}

	w.setCSP()

	for key, values := range w.result {
		// Do not overwrite existing headers which were set outside the header writer
		if len(rw.Header().Get(key)) > 0 {
			continue
		}

		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
}
