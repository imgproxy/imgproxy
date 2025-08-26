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

// Request is a private struct that builds HTTP response headers for a specific request.
type Request struct {
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
func (w *Writer) NewRequest(originalResponseHeaders http.Header, url string) *Request {
	return &Request{
		writer:                  w,
		originalResponseHeaders: originalResponseHeaders,
		url:                     url,
		result:                  make(http.Header),
		maxAge:                  -1,
	}
}

// SetIsFallbackImage sets the Fallback-Image header to
// indicate that the fallback image was used.
func (r *Request) SetIsFallbackImage() {
	// We set maxAge to FallbackImageTTL if it's explicitly passed
	if r.writer.config.FallbackImageTTL < 0 {
		return
	}

	// However, we should not overwrite existing value if set (or greater than ours)
	if r.maxAge < 0 || r.maxAge > r.writer.config.FallbackImageTTL {
		r.maxAge = r.writer.config.FallbackImageTTL
	}
}

// SetExpires sets the TTL from time
func (r *Request) SetExpires(expires *time.Time) {
	if expires == nil {
		return
	}

	// Convert current maxAge to time
	currentMaxAgeTime := time.Now().Add(time.Duration(r.maxAge) * time.Second)

	// If maxAge outlives expires or was not set, we'll use expires as maxAge.
	if r.maxAge < 0 || expires.Before(currentMaxAgeTime) {
		r.maxAge = min(r.writer.config.DefaultTTL, max(0, int(time.Until(*expires).Seconds())))
	}
}

// SetLastModified sets the Last-Modified header from request
func (r *Request) SetLastModified() {
	if !r.writer.config.LastModifiedEnabled {
		return
	}

	val := r.originalResponseHeaders.Get(httpheaders.LastModified)
	if len(val) == 0 {
		return
	}

	r.result.Set(httpheaders.LastModified, val)
}

// SetVary sets the Vary header
func (r *Request) SetVary() {
	if len(r.writer.varyValue) > 0 {
		r.result.Set(httpheaders.Vary, r.writer.varyValue)
	}
}

// SetContentDisposition sets the Content-Disposition header, passthrough to ContentDispositionValue
func (r *Request) SetContentDisposition(originURL, filename, ext, contentType string, returnAttachment bool) {
	value := httpheaders.ContentDispositionValue(
		originURL,
		filename,
		ext,
		contentType,
		returnAttachment,
	)

	if value != "" {
		r.result.Set(httpheaders.ContentDisposition, value)
	}
}

// Passthrough copies specified headers from the original response headers to the response headers.
func (r *Request) Passthrough(only []string) {
	httpheaders.Copy(r.originalResponseHeaders, r.result, only)
}

// CopyFrom copies specified headers from the headers object. Please note that
// all the past operations may overwrite those values.
func (r *Request) CopyFrom(headers http.Header, only []string) {
	httpheaders.Copy(headers, r.result, only)
}

// SetContentLength sets the Content-Length header
func (r *Request) SetContentLength(contentLength int) {
	if contentLength < 0 {
		return
	}

	r.result.Set(httpheaders.ContentLength, strconv.Itoa(contentLength))
}

// SetContentType sets the Content-Type header
func (r *Request) SetContentType(mime string) {
	r.result.Set(httpheaders.ContentType, mime)
}

// writeCanonical sets the Link header with the canonical URL.
// It is mandatory for any response if enabled in the configuration.
func (r *Request) SetCanonical() {
	if !r.writer.config.SetCanonicalHeader {
		return
	}

	if strings.HasPrefix(r.url, "https://") || strings.HasPrefix(r.url, "http://") {
		value := fmt.Sprintf(`<%s>; rel="canonical"`, r.url)
		r.result.Set(httpheaders.Link, value)
	}
}

// setCacheControl sets the Cache-Control header with the specified value.
func (r *Request) setCacheControl(value int) bool {
	if value <= 0 {
		return false
	}

	r.result.Set(httpheaders.CacheControl, fmt.Sprintf("max-age=%d, public", value))
	return true
}

// setCacheControlNoCache sets the Cache-Control header to no-cache (default).
func (r *Request) setCacheControlNoCache() {
	r.result.Set(httpheaders.CacheControl, "no-cache")
}

// setCacheControlPassthrough sets the Cache-Control header from the request
// if passthrough is enabled in the configuration.
func (r *Request) setCacheControlPassthrough() bool {
	if !r.writer.config.CacheControlPassthrough || r.maxAge > 0 {
		return false
	}

	if val := r.originalResponseHeaders.Get(httpheaders.CacheControl); val != "" {
		r.result.Set(httpheaders.CacheControl, val)
		return true
	}

	if val := r.originalResponseHeaders.Get(httpheaders.Expires); val != "" {
		if t, err := time.Parse(http.TimeFormat, val); err == nil {
			maxAge := max(0, int(time.Until(t).Seconds()))
			return r.setCacheControl(maxAge)
		}
	}

	return false
}

// setCSP sets the Content-Security-Policy header to prevent script execution.
func (r *Request) setCSP() {
	r.result.Set(httpheaders.ContentSecurityPolicy, "script-src 'none'")
}

// SetETag copies the ETag header from the original response headers to the result headers.
func (r *Request) SetETag() {
	if !r.writer.config.ETagEnabled {
		return
	}

	etag := r.originalResponseHeaders.Get(httpheaders.Etag)
	if len(etag) == 0 {
		return
	}

	r.result.Set(httpheaders.Etag, etag)
}

// Write writes the headers to the response writer. It does not overwrite
// target headers, which were set outside the header writer.
func (r *Request) Write(rw http.ResponseWriter) {
	// Then, let's try to set Cache-Control using priority order
	switch {
	case r.setCacheControl(r.maxAge): // First, try set explicit
	case r.setCacheControlPassthrough(): // Try to pick up from request headers
	case r.setCacheControl(r.writer.config.DefaultTTL): // Fallback to default value
	default:
		r.setCacheControlNoCache() // By default we use no-cache
	}

	r.setCSP()

	// Copy all headers to the response without overwriting existing ones
	httpheaders.CopyAll(r.result, rw.Header(), false)
}
