package common

import (
	"net/url"
	"strings"
)

func GetBucketAndKey(u *url.URL) (bucket, key string) {
	bucket = u.Host

	// We can't use u.Path here because `url.Parse` unescapes the original URL's path.
	// So we have to use `u.RawPath` if it's available.
	// If it is not available, then `u.EscapedPath()` is the same as the original URL's path
	// before `url.Parse`.
	// See: https://cs.opensource.google/go/go/+/refs/tags/go1.22.4:src/net/url/url.go;l=680
	if len(u.RawPath) > 0 {
		key = u.RawPath
	} else {
		key = u.EscapedPath()
	}

	key = strings.TrimLeft(key, "/")

	return
}
