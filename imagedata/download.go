package imagedata

import (
	"context"
	"net/http"
)

var (
	// For tests. This needs to move to fetcher once we will have a way to isolate
	// the fetcher in tests.
	redirectAllRequestsTo string
)

type DownloadOptions struct {
	Header           http.Header
	CookieJar        http.CookieJar
	MaxSrcFileSize   int
	DownloadFinished context.CancelFunc
}

// RedirectAllRequestsTo TODO: get rid of this global variable
func RedirectAllRequestsTo(u string) {
	redirectAllRequestsTo = u
}

func StopRedirectingRequests() {
	redirectAllRequestsTo = ""
}
