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

// Writer is a struct that creates header writer factories.
type Writer struct {
	config    *Config
	varyValue string
}

// writer is a private struct that builds HTTP response headers for a specific request.
type writer struct {
	writer                  *Writer
	originalResponseHeaders http.Header // Original response headers
	result                  http.Header // Headers to be written to the response
	maxAge                  int         // Current max age for Cache-Control header
	url                     string      // URL of the request, used for canonical header
}

// New creates a new header writer factory with the provided config.
func New(config *Config) (*Writer, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	vary := make([]string, 0)

	if config.SetVaryAccept {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	varyValue := strings.Join(vary, ", ")

	return &Writer{
		config:    config,
		varyValue: varyValue,
	}, nil
}

// NewRequest creates a new header writer instance for a specific request with the provided origin headers and URL.
func (w *Writer) NewRequest(originalResponseHeaders http.Header, url string) *writer {
	return &writer{
		writer:                  w,
		originalResponseHeaders: originalResponseHeaders,
		url:                     url,
		result:                  make(http.Header),
		maxAge:                  -1,
	}
}

// SetIsFallbackImage sets the Fallback-Image header to
// indicate that the fallback image was used.
func (w *writer) SetIsFallbackImage() {
	// We set maxAge to FallbackImageTTL if it's explicitly passed
	if w.writer.config.FallbackImageTTL < 0 {
		return
	}

	// However, we should not overwrite existing value if set (or greater than ours)
	if w.maxAge < 0 || w.maxAge > w.writer.config.FallbackImageTTL {
		w.maxAge = w.writer.config.FallbackImageTTL
	}
}

// SetExpires sets the TTL from time
func (w *writer) SetExpires(expires *time.Time) {
	if expires == nil {
		return
	}

	// Convert current maxAge to time
	currentMaxAgeTime := time.Now().Add(time.Duration(w.maxAge) * time.Second)

	// If maxAge outlives expires or was not set, we'll use expires as maxAge.
	if w.maxAge < 0 || expires.Before(currentMaxAgeTime) {
		w.maxAge = min(w.writer.config.DefaultTTL, max(0, int(time.Until(*expires).Seconds())))
	}
}

// SetLastModified sets the Last-Modified header from request
func (w *writer) SetLastModified() {
	if !w.writer.config.LastModifiedEnabled {
		return
	}

	val := w.originalResponseHeaders.Get(httpheaders.LastModified)
	if len(val) == 0 {
		return
	}

	w.result.Set(httpheaders.LastModified, val)
}

// SetVary sets the Vary header
func (w *writer) SetVary() {
	if len(w.writer.varyValue) > 0 {
		w.result.Set(httpheaders.Vary, w.writer.varyValue)
	}
}

// Passthrough copies specified headers from the original response headers to the response headers.
func (w *writer) Passthrough(only []string) {
	for _, key := range only {
		values := w.originalResponseHeaders.Values(key)

		for _, value := range values {
			w.result.Add(key, value)
		}
	}
}

// CopyFrom copies specified headers from the headers object. Please note that
// all the past operations may overwrite those values.
func (w *writer) CopyFrom(headers http.Header, only []string) {
	for _, key := range only {
		values := headers.Values(key)

		for _, value := range values {
			w.result.Add(key, value)
		}
	}
}

// SetContentLength sets the Content-Length header
func (w *writer) SetContentLength(contentLength int) {
	if contentLength < 0 {
		return
	}

	w.result.Set(httpheaders.ContentLength, strconv.Itoa(contentLength))
}

// SetContentType sets the Content-Type header
func (w *writer) SetContentType(mime string) {
	w.result.Set(httpheaders.ContentType, mime)
}

// writeCanonical sets the Link header with the canonical URL.
// It is mandatory for any response if enabled in the configuration.
func (w *writer) SetCanonical() {
	if !w.writer.config.SetCanonicalHeader {
		return
	}

	if strings.HasPrefix(w.url, "https://") || strings.HasPrefix(w.url, "http://") {
		value := fmt.Sprintf(`<%s>; rel="canonical"`, w.url)
		w.result.Set(httpheaders.Link, value)
	}
}

// setCacheControl sets the Cache-Control header with the specified value.
func (w *writer) setCacheControl(value int) bool {
	if value <= 0 {
		return false
	}

	w.result.Set(httpheaders.CacheControl, fmt.Sprintf("max-age=%d, public", value))
	return true
}

// setCacheControlNoCache sets the Cache-Control header to no-cache (default).
func (w *writer) setCacheControlNoCache() {
	w.result.Set(httpheaders.CacheControl, "no-cache")
}

// setCacheControlPassthrough sets the Cache-Control header from the request
// if passthrough is enabled in the configuration.
func (w *writer) setCacheControlPassthrough() bool {
	if !w.writer.config.CacheControlPassthrough || w.maxAge > 0 {
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
func (w *writer) setCSP() {
	w.result.Set(httpheaders.ContentSecurityPolicy, "script-src 'none'")
}

// Write writes the headers to the response writer. It does not overwrite
// target headers, which were set outside the header writer.
func (w *writer) Write(rw http.ResponseWriter) {
	// Then, let's try to set Cache-Control using priority order
	switch {
	case w.setCacheControl(w.maxAge): // First, try set explicit
	case w.setCacheControlPassthrough(): // Try to pick up from request headers
	case w.setCacheControl(w.writer.config.DefaultTTL): // Fallback to default value
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
