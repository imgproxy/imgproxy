package common

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// IsNotModified returns true if a file was not modified according to
// request/response headers
func IsNotModified(reqHeader http.Header, respHeader http.Header) bool {
	// Etag has higher priority than Last-Modified, so check it first
	if have, matches := httpheaders.CompareEtag(reqHeader, respHeader); have {
		return matches
	}

	// If Etag is not present, check Last-Modified
	if have, matches := httpheaders.CompareLastModified(reqHeader, respHeader); have {
		return matches
	}

	return false
}
