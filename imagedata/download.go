package imagedata

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"syscall"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/security"

	azureTransport "github.com/imgproxy/imgproxy/v3/transport/azure"
	fsTransport "github.com/imgproxy/imgproxy/v3/transport/fs"
	gcsTransport "github.com/imgproxy/imgproxy/v3/transport/gcs"
	s3Transport "github.com/imgproxy/imgproxy/v3/transport/s3"
	swiftTransport "github.com/imgproxy/imgproxy/v3/transport/swift"
)

var (
	downloadClient *http.Client

	enabledSchemes = map[string]struct{}{
		"http":  {},
		"https": {},
	}

	imageHeadersToStore = []string{
		"Cache-Control",
		"Expires",
		"ETag",
	}

	// For tests
	redirectAllRequestsTo string
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

type DownloadOptions struct {
	Header    http.Header
	CookieJar *cookiejar.Jar
}

type ErrorNotModified struct {
	Message string
	Headers map[string]string
}

func (e *ErrorNotModified) Error() string {
	return e.Message
}

func initDownloading() error {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
			Control: func(network, address string, c syscall.RawConn) error {
				return security.VerifySourceNetwork(address)
			},
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   config.Concurrency + 1,
		IdleConnTimeout:       time.Duration(config.ClientKeepAliveTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		DisableCompression:    true,
	}

	if config.ClientKeepAliveTimeout <= 0 {
		transport.MaxIdleConnsPerHost = -1
		transport.DisableKeepAlives = true
	}

	if config.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	registerProtocol := func(scheme string, rt http.RoundTripper) {
		transport.RegisterProtocol(scheme, rt)
		enabledSchemes[scheme] = struct{}{}
	}

	if config.LocalFileSystemRoot != "" {
		registerProtocol("local", fsTransport.New())
	}

	if config.S3Enabled {
		if t, err := s3Transport.New(); err != nil {
			return err
		} else {
			registerProtocol("s3", t)
		}
	}

	if config.GCSEnabled {
		if t, err := gcsTransport.New(); err != nil {
			return err
		} else {
			registerProtocol("gs", t)
		}
	}

	if config.ABSEnabled {
		if t, err := azureTransport.New(); err != nil {
			return err
		} else {
			registerProtocol("abs", t)
		}
	}

	if config.SwiftEnabled {
		if t, err := swiftTransport.New(); err != nil {
			return err
		} else {
			registerProtocol("swift", t)
		}
	}

	downloadClient = &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirects := len(via)
			if redirects >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", redirects)
			}
			return nil
		},
	}

	return nil
}

func headersToStore(res *http.Response) map[string]string {
	m := make(map[string]string)

	for _, h := range imageHeadersToStore {
		if val := res.Header.Get(h); len(val) != 0 {
			m[h] = val
		}
	}

	return m
}

func BuildImageRequest(ctx context.Context, imageURL string, header http.Header, jar *cookiejar.Jar) (*http.Request, context.CancelFunc, error) {
	reqCtx, reqCancel := context.WithTimeout(ctx, time.Duration(config.DownloadTimeout)*time.Second)

	req, err := http.NewRequestWithContext(reqCtx, "GET", imageURL, nil)
	if err != nil {
		reqCancel()
		return nil, func() {}, ierrors.New(404, err.Error(), msgSourceImageIsUnreachable)
	}

	if _, ok := enabledSchemes[req.URL.Scheme]; !ok {
		reqCancel()
		return nil, func() {}, ierrors.New(
			404,
			fmt.Sprintf("Unknown scheme: %s", req.URL.Scheme),
			msgSourceImageIsUnreachable,
		)
	}

	if jar != nil {
		for _, cookie := range jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	req.Header.Set("User-Agent", config.UserAgent)

	for k, v := range header {
		if len(v) > 0 {
			req.Header.Set(k, v[0])
		}
	}

	return req, reqCancel, nil
}

func SendRequest(req *http.Request) (*http.Response, error) {
	res, err := downloadClient.Do(req)
	if err != nil {
		return nil, wrapError(err)
	}

	return res, nil
}

func requestImage(ctx context.Context, imageURL string, opts DownloadOptions) (*http.Response, context.CancelFunc, error) {
	req, reqCancel, err := BuildImageRequest(ctx, imageURL, opts.Header, opts.CookieJar)
	if err != nil {
		reqCancel()
		return nil, func() {}, err
	}

	res, err := SendRequest(req)
	if err != nil {
		reqCancel()
		return nil, func() {}, err
	}

	if res.StatusCode == http.StatusNotModified {
		res.Body.Close()
		reqCancel()
		return nil, func() {}, &ErrorNotModified{Message: "Not Modified", Headers: headersToStore(res)}
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		reqCancel()

		status := 404
		if res.StatusCode >= 500 {
			status = 500
		}

		msg := fmt.Sprintf("Status: %d; %s", res.StatusCode, string(body))
		return nil, func() {}, ierrors.New(status, msg, msgSourceImageIsUnreachable)
	}

	return res, reqCancel, nil
}

func download(ctx context.Context, imageURL string, opts DownloadOptions, secopts security.Options) (*ImageData, error) {
	// We use this for testing
	if len(redirectAllRequestsTo) > 0 {
		imageURL = redirectAllRequestsTo
	}

	res, reqCancel, err := requestImage(ctx, imageURL, opts)
	defer reqCancel()

	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	body := res.Body
	contentLength := int(res.ContentLength)

	if res.Header.Get("Content-Encoding") == "gzip" {
		gzipBody, errGzip := gzip.NewReader(res.Body)
		if gzipBody != nil {
			defer gzipBody.Close()
		}
		if errGzip != nil {
			return nil, err
		}
		body = gzipBody
		contentLength = 0
	}

	imgdata, err := readAndCheckImage(body, contentLength, secopts)
	if err != nil {
		return nil, ierrors.Wrap(err, 0)
	}

	imgdata.Headers = headersToStore(res)

	return imgdata, nil
}

func RedirectAllRequestsTo(u string) {
	redirectAllRequestsTo = u
}

func StopRedirectingRequests() {
	redirectAllRequestsTo = ""
}
