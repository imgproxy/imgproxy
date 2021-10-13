package imagedata

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"

	azureTransport "github.com/imgproxy/imgproxy/v3/transport/azure"
	fsTransport "github.com/imgproxy/imgproxy/v3/transport/fs"
	gcsTransport "github.com/imgproxy/imgproxy/v3/transport/gcs"
	s3Transport "github.com/imgproxy/imgproxy/v3/transport/s3"
)

var (
	downloadClient *http.Client

	imageHeadersToStore = []string{
		"Cache-Control",
		"Expires",
		"ETag",
	}

	// For tests
	redirectAllRequestsTo string
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

type ErrorNotModified struct {
	Message string
	Headers map[string]string
}

func (e *ErrorNotModified) Error() string {
	return e.Message
}

func initDownloading() error {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        config.Concurrency,
		MaxIdleConnsPerHost: config.Concurrency,
		DisableCompression:  true,
		DialContext:         (&net.Dialer{KeepAlive: 600 * time.Second}).DialContext,
	}

	if config.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if config.LocalFileSystemRoot != "" {
		transport.RegisterProtocol("local", fsTransport.New())
	}

	if config.S3Enabled {
		if t, err := s3Transport.New(); err != nil {
			return err
		} else {
			transport.RegisterProtocol("s3", t)
		}
	}

	if config.GCSEnabled {
		if t, err := gcsTransport.New(); err != nil {
			return err
		} else {
			transport.RegisterProtocol("gs", t)
		}
	}

	if config.ABSEnabled {
		if t, err := azureTransport.New(); err != nil {
			return err
		} else {
			transport.RegisterProtocol("abs", t)
		}
	}

	downloadClient = &http.Client{
		Timeout:   time.Duration(config.DownloadTimeout) * time.Second,
		Transport: transport,
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

func requestImage(imageURL string, header http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, ierrors.New(404, err.Error(), msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	req.Header.Set("User-Agent", config.UserAgent)

	for k, v := range header {
		if len(v) > 0 {
			req.Header.Set(k, v[0])
		}
	}

	res, err := downloadClient.Do(req)
	if err != nil {
		return res, ierrors.New(404, checkTimeoutErr(err).Error(), msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	if res.StatusCode == http.StatusNotModified {
		return nil, &ErrorNotModified{Message: "Not Modified", Headers: headersToStore(res)}
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		msg := fmt.Sprintf("Status: %d; %s", res.StatusCode, string(body))
		return res, ierrors.New(404, msg, msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	return res, nil
}

func download(imageURL string, header http.Header) (*ImageData, error) {
	// We use this for testing
	if len(redirectAllRequestsTo) > 0 {
		imageURL = redirectAllRequestsTo
	}

	res, err := requestImage(imageURL, header)
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

	imgdata, err := readAndCheckImage(body, contentLength)
	if err != nil {
		return nil, err
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
