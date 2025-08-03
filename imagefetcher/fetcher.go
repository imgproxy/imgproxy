// imagefetcher is responsible for downloading images using HTTP requests through various protocols
// defined in transport package
package imagefetcher

import (
	"context"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
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
	transport    *transport.Transport // Transport used for making HTTP requests
	maxRedirects int                  // Maximum number of redirects allowed
}

// NewFetcher creates a new ImageFetcher with the provided transport
func NewFetcher(transport *transport.Transport, maxRedirects int) (*Fetcher, error) {
	return &Fetcher{transport, maxRedirects}, nil
}

// checkRedirect is a method that checks if the number of redirects exceeds the maximum allowed
func (f *Fetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	redirects := len(via)
	if redirects >= f.maxRedirects {
		return newTooManyRedirectsError(redirects)
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

// BuildRequest creates a new ImageFetcherRequest with the provided context, URL, headers, and cookie jar
func (f *Fetcher) BuildRequest(ctx context.Context, url string, header http.Header, jar http.CookieJar) (*Request, error) {
	url = common.EscapeURL(url)

	// Set request timeout and get cancel function
	ctx, cancel := context.WithTimeout(ctx, time.Duration(config.DownloadTimeout)*time.Second)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		cancel()
		return nil, newRequestError(err)
	}

	// Check if the URL scheme is supported
	if !f.transport.IsProtocolRegistered(req.URL.Scheme) {
		cancel()
		return nil, newRequestSchemeError(req.URL.Scheme)
	}

	// Add cookies from the jar to the request (if any)
	if jar != nil {
		for _, cookie := range jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	// Set user agent header
	req.Header.Set(httpheaders.UserAgent, config.UserAgent)

	// Set headers
	for k, v := range header {
		if len(v) > 0 {
			req.Header.Set(k, v[0])
		}
	}

	return &Request{f, req, cancel}, nil
}
