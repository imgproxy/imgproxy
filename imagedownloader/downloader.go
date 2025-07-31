// Package imagedownloader provides a shared method for downloading any
// images within imgproxy.
package imagedownloader

import (
	"context"
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/transport"
)

var (
	// Global downloader instance
	Fetcher *imagefetcher.Fetcher
	D       *Downloader

	// For tests
	redirectAllRequestsTo string
)

type DownloadOptions struct {
	Header    http.Header
	CookieJar http.CookieJar
}

// Downloader is responsible for downloading images and converting them to ImageData
type Downloader struct {
	fetcher *imagefetcher.Fetcher
}

// NewDownloader creates a new Downloader with the provided fetcher and config
func NewDownloader(fetcher *imagefetcher.Fetcher) *Downloader {
	return &Downloader{
		fetcher: fetcher,
	}
}

// Init initializes the global downloader
func InitGlobalDownloader() error {
	ts, err := transport.NewTransport()
	if err != nil {
		return err
	}

	Fetcher, err := imagefetcher.NewFetcher(ts, imagefetcher.NewConfigFromEnv())
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create image fetcher"))
	}

	D = NewDownloader(Fetcher)

	return nil
}

// Download downloads an image from the given URL and returns ImageData
func (d *Downloader) Download(ctx context.Context, imageURL string, opts DownloadOptions, secopts security.Options) (imagedatanew.ImageData, error) {
	// We use this for testing
	if len(redirectAllRequestsTo) > 0 {
		imageURL = redirectAllRequestsTo
	}

	req, err := d.fetcher.BuildRequest(ctx, imageURL, opts.Header, opts.CookieJar)
	if err != nil {
		defer req.Cancel()
		return nil, err
	}

	res, err := req.FetchImage()
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return nil, err
	}

	// Create factory with the provided security options for this request
	imgdata, err := imagedatanew.NewFromResponse(res, secopts)
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return nil, err
	}

	return imgdata, nil
}

// DownloadWithDesc downloads an image from the given URL, gives error a description context and returns ImageData
func (d *Downloader) DownloadWithDesc(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (imagedatanew.ImageData, error) {
	i, err := d.Download(ctx, imageURL, opts, secopts)

	if err != nil {
		return nil, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("Can't download %s", desc)),
		)
	}

	return i, err
}

// Download downloads an image using the global downloader.
// NOTE: This method uses globalDownloader instance. In the future, this will
// be replaced with an instance everywhere.
func Download(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (imagedatanew.ImageData, error) {
	return D.DownloadWithDesc(ctx, imageURL, desc, opts, secopts)
}

// RedirectAllRequestsTo redirects all requests to the given URL (for testing)
func RedirectAllRequestsTo(u string) {
	redirectAllRequestsTo = u
}

// StopRedirectingRequests stops redirecting requests (for testing)
func StopRedirectingRequests() {
	redirectAllRequestsTo = ""
}
