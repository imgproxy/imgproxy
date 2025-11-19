package common

import (
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// IsNotModified returns true if a file was not modified according to
// request/response headers
func IsNotModified(reqHeader http.Header, header http.Header) bool {
	etag := header.Get(httpheaders.Etag)
	ifNoneMatch := reqHeader.Get(httpheaders.IfNoneMatch)

	if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
		return true
	}

	lastModifiedRaw := header.Get(httpheaders.LastModified)
	if len(lastModifiedRaw) == 0 {
		return false
	}

	ifModifiedSinceRaw := reqHeader.Get(httpheaders.IfModifiedSince)
	if len(ifModifiedSinceRaw) == 0 {
		return false
	}

	lastModified, err := time.Parse(http.TimeFormat, lastModifiedRaw)
	if err != nil {
		return false
	}

	ifModifiedSince, err := time.Parse(http.TimeFormat, ifModifiedSinceRaw)
	if err != nil {
		return false
	}

	if !ifModifiedSince.Before(lastModified) {
		return true
	}

	return false
}
