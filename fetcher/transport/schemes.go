package transport

import "net/url"

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"

	SchemeLocal = "local"
	SchemeS3    = "s3"
	SchemeGCS   = "gs"
	SchemeAzure = "abs"
	SchemeSwift = "swift"
)

func IsUrlSchemeHTTP(u *url.URL) bool {
	return u.Scheme == SchemeHTTP || u.Scheme == SchemeHTTPS
}
