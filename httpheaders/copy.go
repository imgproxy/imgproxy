package httpheaders

import (
	"net/http"
)

// Copy copies specified headers from one header to another.
func Copy(from, to http.Header, only []string) {
	for _, key := range only {
		key = http.CanonicalHeaderKey(key)
		if values := from[key]; len(values) > 0 {
			to[key] = append([]string(nil), values...)
		}
	}
}

// CopyAll copies all headers from one header to another.
func CopyAll(from, to http.Header, overwrite bool) {
	for key, values := range from {
		// Keys in http.Header are already canonicalized, so no need for http.CanonicalHeaderKey here
		if !overwrite && len(to.Values(key)) > 0 {
			continue
		}

		if len(values) > 0 {
			to[key] = append([]string(nil), values...)
		}
	}
}

// CopyFromRequest copies specified headers from the http.Request to the provided header.
func CopyFromRequest(req *http.Request, header http.Header, only []string) {
	for _, key := range only {
		key = http.CanonicalHeaderKey(key)

		if key == Host {
			header.Set(key, req.Host)
			continue
		}

		if values := req.Header[key]; len(values) > 0 {
			header[key] = append([]string(nil), values...)
		}
	}
}

// CopyToRequest copies headers from the provided header to the http.Request.
func CopyToRequest(header http.Header, req *http.Request) {
	for key, values := range header {
		if len(values) == 0 {
			continue
		}

		// Keys in http.Header are already canonicalized, so no need for http.CanonicalHeaderKey here
		if key == Host {
			req.Host = values[0]
		} else {
			req.Header[key] = append([]string(nil), values...)
		}
	}
}
