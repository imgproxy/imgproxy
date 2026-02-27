package httpheaders

import (
	"net/http"
	"strings"
)

// CompareEtag compares the ETag from the response header
// with the If-None-Match header from the request.
// It returns two boolean values:
//
// - The first boolean indicates that both ETag and If-None-Match headers are present
// - The second boolean indicates that the ETag matches the If-None-Match value,
// meaning the resource has not been modified.
func CompareEtag(reqHeader, respHeader http.Header) (bool, bool) {
	ifNoneMatch := reqHeader.Get(IfNoneMatch)
	if ifNoneMatch == "" {
		return false, false
	}

	return CompareEtagValue(ifNoneMatch, respHeader)
}

// CompareEtagValue same method as above, but accepts the raw If-None-Match value
// instead of the request header.
func CompareEtagValue(ifNoneMatch string, respHeader http.Header) (bool, bool) {
	etag := respHeader.Get(Etag)
	if etag == "" {
		return false, false
	}

	return true, UnquoteEtag(etag) == UnquoteEtag(ifNoneMatch)
}

// UnquoteEtag removes quotes from the ETag value if they are present.
// It also removes the weak ETag prefix "W/" if it is present.
func UnquoteEtag(etag string) string {
	etag = strings.TrimPrefix(etag, "W/")
	etag = strings.Trim(etag, `"`)
	return etag
}

// CompareLastModified compares the Last-Modified header from the response
// with the If-Modified-Since header from the request.
// It returns two boolean values:
//
// - The first boolean indicates that both Last-Modified and If-Modified-Since headers are present
// - The second boolean indicates that the resource has not been modified since the time specified
// in the If-Modified-Since header.
func CompareLastModified(reqHeader, respHeader http.Header) (bool, bool) {
	ifModifiedSinceStr := reqHeader.Get(IfModifiedSince)
	if ifModifiedSinceStr == "" {
		return false, false
	}

	return CompareLastModifiedValue(ifModifiedSinceStr, respHeader)
}

// CompareLastModifiedValue same as above, but accepts the raw If-Modified-Since value
func CompareLastModifiedValue(ifModifiedSinceStr string, respHeader http.Header) (bool, bool) {
	ifModifiedSince, err := http.ParseTime(ifModifiedSinceStr)
	if err != nil {
		return false, false
	}

	lastModifiedStr := respHeader.Get(LastModified)
	if lastModifiedStr == "" {
		return false, false
	}

	lastModified, err := http.ParseTime(lastModifiedStr)
	if err != nil {
		return false, false
	}

	return true, !lastModified.After(ifModifiedSince)
}
