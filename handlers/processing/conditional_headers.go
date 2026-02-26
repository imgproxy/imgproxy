package processing

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

type ConditionalHeaders struct {
	config *Config // just a shortcut to the config

	ifModifiedSince string // raw value of the If-Modified-Since header from the user request
	ifNoneMatch     string // raw value of the If-None-Match header from the user request
	originHeaders   http.Header
}

// NewConditionalHeadersFromRequest creates a new ConditionalHeaders instance from the given request.
func NewConditionalHeadersFromRequest(c *Config, req *http.Request) *ConditionalHeaders {
	ifModifiedSince := req.Header.Get(httpheaders.IfModifiedSince)
	ifNoneMatch := req.Header.Get(httpheaders.IfNoneMatch)

	return &ConditionalHeaders{
		config:          c,
		ifModifiedSince: ifModifiedSince,
		ifNoneMatch:     ifNoneMatch,
		originHeaders:   nil,
	}
}

// SetOriginHeaders sets the origin headers for the request.
func (c *ConditionalHeaders) SetOriginHeaders(h http.Header) {
	c.originHeaders = h
}

// InjectImageRequestHeaders injects conditional headers into the source
// image request if needed.
func (c *ConditionalHeaders) InjectImageRequestHeaders(imageReqHeaders http.Header) {
	var abort bool

	etag, abort := c.computeEtag()
	if abort {
		return
	}

	ifModifiedSince := c.computeIfModifiedSince()

	if len(ifModifiedSince) > 0 {
		imageReqHeaders.Set(httpheaders.IfModifiedSince, ifModifiedSince)
	}

	if len(etag) > 0 {
		imageReqHeaders.Set(httpheaders.IfNoneMatch, etag)
	}
}

// InjectUserResponseHeaders injects conditional headers into the user response
func (c *ConditionalHeaders) InjectUserResponseHeaders(rw http.ResponseWriter) {
	c.injectLastModifiedHeader(rw)
	c.injectEtagHeader(rw)
}

// computeIfModifiedSince determines whether the If-Modified-Since header should
// be sent to the source image server. It returns value to be set (if any)
func (c *ConditionalHeaders) computeIfModifiedSince() string {
	// If the feature is disabled or no header is present, we shouldn't
	// send the header, but that should not affect other headers
	if !c.config.LastModifiedEnabled || len(c.ifModifiedSince) == 0 {
		return ""
	}

	// No buster is set: we should send the header as is
	if c.config.LastModifiedBuster.IsZero() {
		return c.ifModifiedSince
	}

	// Parse the header
	ifModifiedSince, err := http.ParseTime(c.ifModifiedSince)

	// Header has invalid format, or
	// the buster is set, and header is older than the buster
	if err != nil || !c.config.LastModifiedBuster.Before(ifModifiedSince) {
		return ""
	}

	// Otherwise no conditional headers should be sent at all
	return c.ifModifiedSince
}

// computeEtag determines whether the If-None-Match header should be sent to the
// source image server. It returns etag value to be set and boolean indicating
// whether the conditional headers should be sent at all.
func (c *ConditionalHeaders) computeEtag() (string, bool) {
	// If the feature is disabled or no header is present,
	// we shouldn't send the header at all, but it should not affect other headers
	if !c.config.ETagEnabled || len(c.ifNoneMatch) == 0 {
		return "", false
	}

	// If etag buster is not set, we should send the header as is if present
	if len(c.config.ETagBuster) == 0 {
		return c.ifNoneMatch, false
	}

	// Unquote and remove /W
	ifNoneMatch := httpheaders.UnquoteEtag(c.ifNoneMatch)

	// We expect that incoming ETag header has the buster
	rest, busterFound := strings.CutPrefix(ifNoneMatch, c.config.ETagBuster+"/")
	if !busterFound {
		return "", true // do not send any conditional headers otherwise (???)
	}

	// Parse the rest of the header as base64-encoded string, if it fails,
	// we should not send any conditional headers (invalid etag)
	etag, err := base64.RawURLEncoding.DecodeString(rest)
	if err != nil {
		return "", true // do not send any conditional headers otherwise
	}

	// Quotes will be encoded into etag
	return string(etag), false
}

// injectLastModifiedHeader injects the Last-Modified header into the user response
func (c *ConditionalHeaders) injectLastModifiedHeader(rw http.ResponseWriter) {
	// If the feature is disabled, we shouldn't send the header at all
	if !c.config.LastModifiedEnabled {
		return
	}

	// No header is present: nothing to inject
	val := c.originHeaders.Get(httpheaders.LastModified)
	if len(val) == 0 {
		return
	}

	// If the incoming header is older than the buster, we should replace it with the buster
	lastModified, err := http.ParseTime(val)
	if err != nil {
		return // invalid values are not forwarded
	}

	if lastModified.Before(c.config.LastModifiedBuster) {
		val = c.config.LastModifiedBuster.Format(http.TimeFormat)
	}

	// Otherwise, we should just pass the header through
	rw.Header().Set(httpheaders.LastModified, val)
}

// injectEtagHeader injects the ETag header into the user response
func (c *ConditionalHeaders) injectEtagHeader(rw http.ResponseWriter) {
	if !c.config.ETagEnabled {
		return
	}

	etag := c.originHeaders.Get(httpheaders.Etag)
	if len(etag) == 0 {
		return
	}

	if len(c.config.ETagBuster) > 0 {
		etag = `"` + c.config.ETagBuster + "/" + base64.RawURLEncoding.EncodeToString([]byte(etag)) + `"`
	}

	if len(etag) > 0 {
		rw.Header().Set(httpheaders.Etag, etag)
	}
}
