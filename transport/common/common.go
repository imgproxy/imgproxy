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

	// We replaced all `%` with `%25` in `imagedata.BuildImageRequest` to prevent parsing errors.
	// Also, we replaced all `#` with `%23` to prevent cutting off the fragment part.
	// We need to revert these replacements.
	//
	// It's important to revert %23 first because the original URL may also contain %23,
	// and we don't want to mix them up.
	bucket = strings.ReplaceAll(bucket, "%23", "#")
	bucket = strings.ReplaceAll(bucket, "%25", "%")
	key = strings.ReplaceAll(key, "%23", "#")
	key = strings.ReplaceAll(key, "%25", "%")

	return
}
