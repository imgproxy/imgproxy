// headerwriter writes response HTTP headers
package headerwriter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.withmatt.com/httpheaders"
)

const (
	// Content-Disposition header format
	contentDispositionFmt = "%s; filename=\"%s%s\""
)

// Writer is a struct that builds HTTP response headers.
type Writer struct {
	config                  *Config
	originalResponseHeaders http.Header // Original response headers
	res                     http.Header // Headers to be written to the response
	maxAge                  int         // Current max age for Cache-Control header
	url                     string      // URL of the request, used for canonical header
}

// newWriter creates a new HeaderBuilder instance with the provided origin headers and URL
func newWriter(config *Config, originalResponseHeaders http.Header, url string) *Writer {
	return &Writer{
		config:                  config,
		originalResponseHeaders: originalResponseHeaders,
		url:                     url,
		res:                     make(http.Header),
		maxAge:                  -1,
	}
}

// SetMaxAge sets the max-age for the Cache-Control header.
// Overrides any existing max-age value.
func (w *Writer) SetMaxAge(maxAge int) {
	if maxAge > 0 {
		w.maxAge = maxAge
	}
}

// SetIsFallbackImage sets the Fallback-Image header to
// indicate that the fallback image was used.
func (w *Writer) SetIsFallbackImage() {
	w.res.Set("Fallback-Image", "1")
}

// SetMaxAgeTime sets the max-age for the Cache-Control header based
// on the time provided. If time provided is in the past compared
// to the current maxAge value, it will correct maxAge.
func (w *Writer) SetMaxAgeFromExpires(expires *time.Time) {
	if expires == nil {
		return
	}

	// Convert current maxAge to time
	currentMaxAgeTime := time.Now().Add(time.Duration(w.maxAge) * time.Second)

	// If the expires time is in the past compared to the current maxAge time,
	// or if maxAge is not set, we will use the expires time to set the maxAge.
	if w.maxAge < 0 || expires.Before(currentMaxAgeTime) {
		// Get the TTL from the expires time (must not be in the past)
		expiresTTL := max(0, int(time.Until(*expires).Seconds()))

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
	if val == "" {
		return
	}

	w.res.Set(httpheaders.LastModified, val)
}

// SetVary sets the Vary header
func (w *Writer) SetVary() {
	vary := make([]string, 0)

	if w.config.SetVaryAccept {
		vary = append(vary, "Accept")
	}

	if w.config.EnableClientHints {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	varyValue := strings.Join(vary, ", ")

	if varyValue != "" {
		w.res.Set(httpheaders.Vary, varyValue)
	}
}

// Copy copies specified headers from the original response headers to the response headers.
func (w *Writer) Copy(only []string) {
	for _, key := range only {
		values := w.originalResponseHeaders.Values(key)

		for _, value := range values {
			w.res.Add(key, value)
		}
	}
}

// CopyFrom copies specified headers from the headers object. Please note that
// all the past operations may overwrite those values.
func (w *Writer) CopyFrom(headers http.Header, only []string) {
	for _, key := range only {
		values := headers.Values(key)

		for _, value := range values {
			w.res.Add(key, value)
		}
	}
}

// SetContentLength sets the Content-Length header
func (w *Writer) SetContentLength(contentLength int) {
	if contentLength > 0 {
		w.res.Set(httpheaders.ContentLength, strconv.Itoa(contentLength))
	}
}

// SetContentDisposition sets the Content-Disposition header
func (w *Writer) SetContentDisposition(filename, ext string, returnAttachment bool) {
	disposition := "inline"

	if returnAttachment {
		disposition = "attachment"
	}

	value := fmt.Sprintf(contentDispositionFmt, disposition, strings.ReplaceAll(filename, `"`, "%22"), ext)

	w.res.Set(httpheaders.ContentDisposition, value)
}

func (w *Writer) SetContentType(mime string) {
	w.res.Set(httpheaders.ContentType, mime)
}

// writeCanonical sets the Link header with the canonical URL.
// It is mandatory for any response if enabled in the configuration.
func (b *Writer) SetCanonical() {
	if !b.config.SetCanonicalHeader {
		return
	}

	if strings.HasPrefix(b.url, "https://") || strings.HasPrefix(b.url, "http://") {
		value := fmt.Sprintf(`<%s>; rel="canonical"`, b.url)
		b.res.Set(httpheaders.Link, value)
	}
}

// setCacheControlNoCache sets the Cache-Control header to no-cache (default).
func (w *Writer) setCacheControlNoCache() {
	w.res.Set(httpheaders.CacheControl, "no-cache")
}

// setCacheControlMaxAge sets the Cache-Control header with max-age.
func (w *Writer) setCacheControlMaxAge() {
	maxAge := w.maxAge

	if maxAge <= 0 {
		maxAge = w.config.DefaultTTL
	}

	if maxAge > 0 {
		w.res.Set(httpheaders.CacheControl, fmt.Sprintf("max-age=%d, public", maxAge))
	}
}

// setCacheControlPassthrough sets the Cache-Control header from the request
// if passthrough is enabled in the configuration.
func (w *Writer) setCacheControlPassthrough() bool {
	if !w.config.CacheControlPassthrough || w.maxAge > 0 {
		return false
	}

	if val := w.originalResponseHeaders.Get(httpheaders.CacheControl); val != "" {
		w.res.Set(httpheaders.CacheControl, val)
		return true
	}

	if val := w.originalResponseHeaders.Get(httpheaders.Expires); val != "" {
		if t, err := time.Parse(http.TimeFormat, val); err == nil {
			w.maxAge = max(0, int(time.Until(t).Seconds()))
		}
	}

	return false
}

// setCSP sets the Content-Security-Policy header to prevent script execution.
func (w *Writer) setCSP() {
	w.res.Set("Content-Security-Policy", "script-src 'none'")
}

// Write writes the headers to the response writer
func (w *Writer) Write(rw http.ResponseWriter) {
	w.setCacheControlNoCache()

	if !w.setCacheControlPassthrough() {
		w.setCacheControlMaxAge()
	}

	w.setCSP()

	for key, values := range w.res {
		for _, value := range values {
			rw.Header().Add(key, value)
		}
	}
}

// NOTE: WIP

// func (w *HeaderBuilder) SetDebugHeaders() {
// 	if config.EnableDebugHeaders {
// 		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(originData.Data)))
// 		rw.Header().Set("X-Origin-Width", resultData.Headers["X-Origin-Width"])
// 		rw.Header().Set("X-Origin-Height", resultData.Headers["X-Origin-Height"])
// 		rw.Header().Set("X-Result-Width", resultData.Headers["X-Result-Width"])
// 		rw.Header().Set("X-Result-Height", resultData.Headers["X-Result-Height"])
// 	}
// }
