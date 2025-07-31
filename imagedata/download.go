package imagedata

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/transport"
)

var (
	Fetcher *imagefetcher.Fetcher

	// For tests
	redirectAllRequestsTo string
)

type DownloadOptions struct {
	Header    http.Header
	CookieJar http.CookieJar
}

func initDownloading() error {
	ts, err := transport.NewTransport()
	if err != nil {
		return err
	}

	Fetcher, err = imagefetcher.NewFetcher(ts, config.MaxRedirects)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create image fetcher"))
	}

	return nil
}

func download(ctx context.Context, imageURL string, opts DownloadOptions, secopts security.Options) (*ImageData, error) {
	// We use this for testing
	if len(redirectAllRequestsTo) > 0 {
		imageURL = redirectAllRequestsTo
	}

	req, err := Fetcher.BuildRequest(ctx, imageURL, opts.Header, opts.CookieJar)
	if err != nil {
		return nil, err
	}
	defer req.Cancel()

	res, err := req.FetchImage()
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return nil, err
	}

	res, err = security.LimitResponseSize(res, secopts)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	imgdata, err := readAndCheckImage(res.Body, int(res.ContentLength), secopts)
	if err != nil {
		return nil, ierrors.Wrap(err, 0)
	}

	imgdata.headers = res.Header.Clone()

	return imgdata, nil
}

func RedirectAllRequestsTo(u string) {
	redirectAllRequestsTo = u
}

func StopRedirectingRequests() {
	redirectAllRequestsTo = ""
}
