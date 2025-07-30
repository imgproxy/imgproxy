package imagedatanew

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/security"
)

// HTTPOptions defines options for HTTP requests made to fetch images
type HTTPOptions struct {
	Header    http.Header
	CookieJar http.CookieJar
}

// NewFromURL creates a new ImageData making a HTTP request to the specified URL
func NewFromURL(ctx context.Context, fetcher *imagefetcher.Fetcher, url string, opts HTTPOptions, secopts security.Options) (ImageData, error) {
	req, err := fetcher.BuildRequest(ctx, url, opts.Header, opts.CookieJar)
	if err != nil {
		if req != nil {
			defer req.Cancel()
		}
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
	imgdata, err := NewFromResponse(res, secopts)
	if err != nil {
		if res != nil {
			req.Cancel()
			res.Body.Close()
		}
		return nil, err
	}

	return imgdata, nil
}
