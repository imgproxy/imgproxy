package imagedata

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v2/config"
	"github.com/imgproxy/imgproxy/v2/ierrors"

	azureTransport "github.com/imgproxy/imgproxy/v2/transport/azure"
	fsTransport "github.com/imgproxy/imgproxy/v2/transport/fs"
	gcsTransport "github.com/imgproxy/imgproxy/v2/transport/gcs"
	s3Transport "github.com/imgproxy/imgproxy/v2/transport/s3"
)

var (
	downloadClient *http.Client

	imageHeadersToStore = []string{
		"Cache-Control",
		"Expires",
	}
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

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

func requestImage(imageURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, ierrors.New(404, err.Error(), msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	req.Header.Set("User-Agent", config.UserAgent)

	res, err := downloadClient.Do(req)
	if err != nil {
		return res, ierrors.New(404, checkTimeoutErr(err).Error(), msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()

		msg := fmt.Sprintf("Can't download image; Status: %d; %s", res.StatusCode, string(body))
		return res, ierrors.New(404, msg, msgSourceImageIsUnreachable).SetUnexpected(config.ReportDownloadingErrors)
	}

	return res, nil
}

func download(imageURL string) (*ImageData, error) {
	res, err := requestImage(imageURL)
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

	imgdata.Headers = make(map[string]string)
	for _, h := range imageHeadersToStore {
		if val := res.Header.Get(h); len(val) != 0 {
			imgdata.Headers[h] = val
		}
	}

	return imgdata, nil
}
