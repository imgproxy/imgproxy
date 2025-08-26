package imagedata

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/transport"
)

var (
	Fetcher *imagefetcher.Fetcher

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

func DefaultDownloadOptions() DownloadOptions {
	return DownloadOptions{
		Header:           nil,
		CookieJar:        nil,
		MaxSrcFileSize:   config.MaxSrcFileSize,
		DownloadFinished: nil,
	}
}

func initDownloading() error {
	ts, err := transport.NewTransport()
	if err != nil {
		return err
	}

	c, err := imagefetcher.LoadFromEnv(imagefetcher.NewDefaultConfig())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithPrefix("configuration error"))
	}

	Fetcher, err = imagefetcher.NewFetcher(ts, c)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create image fetcher"))
	}

	return nil
}

func RedirectAllRequestsTo(u string) {
	redirectAllRequestsTo = u
}

func StopRedirectingRequests() {
	redirectAllRequestsTo = ""
}
