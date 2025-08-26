// imagefetcher is responsible for downloading images using HTTP requests through various protocols
// defined in transport package
package imagefetcher

import (
	"context"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/transport"
	"github.com/imgproxy/imgproxy/v3/transport/common"
)

const (
	connectionLostError = "client connection lost" // Error message indicating a lost connection
	bounceDelay         = 100 * time.Microsecond   // Delay before retrying a request
)

// Fetcher is a struct that holds the HTTP client and transport for fetching images
type Fetcher struct {
	transport *transport.Transport // Transport used for making HTTP requests
	config    *Config              // Configuration for the image fetcher
}

// NewFetcher creates a new ImageFetcher with the provided transport
func NewFetcher(transport *transport.Transport, config *Config) (*Fetcher, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Fetcher{transport, config}, nil
}

// checkRedirect is a method that checks if the number of redirects exceeds the maximum allowed
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	redirects := len(via)
	if redirects >= f.config.MaxRedirects {
		return newImageTooManyRedirectsError(redirects)
	}
	return nil
}

// newHttpClient returns new HTTP client
func (f *Fetcher) newHttpClient() *http.Client {
	return &http.Client{
		Transport:     f.transport.Transport(), // Connection pool is there
		CheckRedirect: f.checkRedirect,
	}
}

// NewImageFetcherRequest creates a new ImageFetcherRequest with the provided context, URL, headers, and cookie jar
func (f *Fetcher) BuildRequest(ctx context.Context, url string, header http.Header, jar http.CookieJar) (*Request, error) {
	url = common.EscapeURL(url)

	// Set request timeout and get cancel function
	ctx, cancel := context.WithTimeout(ctx, f.config.DownloadTimeout)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return nil, newImageRequestError(err)
	}

	// Check if the URL scheme is supported
	if !f.transport.IsProtocolRegistered(req.URL.Scheme) {
		cancel()
		return nil, newImageRequstSchemeError(req.URL.Scheme)
	}

	// Add cookies from the jar to the request (if any)
	if jar != nil {
		for _, cookie := range jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	// Set user agent header
	req.Header.Set(httpheaders.UserAgent, f.config.UserAgent)

	// Set headers
	httpheaders.CopyToRequest(header, req)

	return &Request{f, req, cancel}, nil
}
