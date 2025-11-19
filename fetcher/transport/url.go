package transport

import (
	"net/url"
	"strings"
)

func EscapeURL(u string) string {
	// Non-http(s) URLs may contain percent symbol outside of the percent-encoded sequences.
	// Parsing such URLs will fail with an error.
	// To prevent this, we replace all percent symbols with %25.
	//
	// Also, such URLs may contain a hash symbol (a fragment identifier) or a question mark
	// (a query string).
	// We replace them with %23 and %3F to make `url.Parse` treat them as a part of the path.
	// Since we already replaced all percent symbols, we won't mix up %23/%3F that were in the
	// original URL and %23/%3F that appeared after the replacement.
	//
	// We will revert these replacements in `GetBucketAndKey`.
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = strings.ReplaceAll(u, "%", "%25")
		u = strings.ReplaceAll(u, "?", "%3F")
		u = strings.ReplaceAll(u, "#", "%23")
	}

	return u
}

// GetBucketAndKey extracts bucket and key from the provided URL.
func GetBucketAndKey(u *url.URL, sep string) (bucket, key, query string) {
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

	// We percent-encoded `%`, `#`, and `?` in `EscapeURL` to prevent parsing errors.
	// Now we need to revert these replacements.
	//
	// It's important to revert %25 last because %23/%3F may appear in the original URL and
	// we don't want to mix them up.
	bucket = strings.ReplaceAll(bucket, "%23", "#")
	bucket = strings.ReplaceAll(bucket, "%3F", "?")
	bucket = strings.ReplaceAll(bucket, "%25", "%")
	key = strings.ReplaceAll(key, "%23", "#")
	key = strings.ReplaceAll(key, "%3F", "?")
	key = strings.ReplaceAll(key, "%25", "%")

	// Cut the query string if it's present.
	// Since we replaced `?` with `%3F` in `EscapeURL`, `url.Parse` will treat query
	// string as a part of the path.
	// Also, query string separator may be different from `?`, so we can't rely on `url.URL.RawQuery`.
	if len(sep) > 0 {
		key, query, _ = strings.Cut(key, sep)
	}

	return
}
