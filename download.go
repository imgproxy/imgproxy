package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mat/besticon/ico"
)

var (
	downloadClient  *http.Client
	imageTypeCtxKey = ctxKey("imageType")
	imageDataCtxKey = ctxKey("imageData")

	errSourceDimensionsTooBig      = newError(422, "Source image dimensions are too big", "Invalid source image")
	errSourceResolutionTooBig      = newError(422, "Source image resolution is too big", "Invalid source image")
	errSourceFileTooBig            = newError(422, "Source image file is too big", "Invalid source image")
	errSourceImageTypeNotSupported = newError(422, "Source image type not supported", "Invalid source image")
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

var downloadBufPool *bufPool

type limitReader struct {
	r    io.ReadCloser
	left int
}

func (lr *limitReader) Read(p []byte) (n int, err error) {
	n, err = lr.r.Read(p)
	lr.left = lr.left - n

	if err == nil && lr.left < 0 {
		err = errSourceFileTooBig
	}

	return
}

func (lr *limitReader) Close() error {
	return lr.r.Close()
}

func initDownloading() {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        conf.Concurrency,
		MaxIdleConnsPerHost: conf.Concurrency,
		DisableCompression:  true,
		Dial:                (&net.Dialer{KeepAlive: 600 * time.Second}).Dial,
	}

	if conf.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if conf.LocalFileSystemRoot != "" {
		transport.RegisterProtocol("local", newFsTransport())
	}

	if conf.S3Enabled {
		transport.RegisterProtocol("s3", newS3Transport())
	}

	if len(conf.GCSKey) > 0 {
		transport.RegisterProtocol("gs", newGCSTransport())
	}

	downloadClient = &http.Client{
		Timeout:   time.Duration(conf.DownloadTimeout) * time.Second,
		Transport: transport,
	}

	downloadBufPool = newBufPool("download", conf.Concurrency, conf.DownloadBufferSize)
}

func checkDimensions(width, height int) error {
	if conf.MaxSrcDimension > 0 && (width > conf.MaxSrcDimension || height > conf.MaxSrcDimension) {
		return errSourceDimensionsTooBig
	}

	if width*height > conf.MaxSrcResolution {
		return errSourceResolutionTooBig
	}

	return nil
}

func checkTypeAndDimensions(r io.Reader) (imageType, error) {
	imgconf, imgtypeStr, err := image.DecodeConfig(r)
	if err == image.ErrFormat {
		return imageTypeUnknown, errSourceImageTypeNotSupported
	}
	if err != nil {
		return imageTypeUnknown, err
	}

	imgtype, imgtypeOk := imageTypes[imgtypeStr]
	if !imgtypeOk || !vipsTypeSupportLoad[imgtype] {
		return imageTypeUnknown, errSourceImageTypeNotSupported
	}

	if err = checkDimensions(imgconf.Width, imgconf.Height); err != nil {
		return imageTypeUnknown, err
	}

	return imgtype, nil
}

func readAndCheckImage(ctx context.Context, res *http.Response) (context.Context, context.CancelFunc, error) {
	var contentLength int

	if res.ContentLength > 0 {
		contentLength = int(res.ContentLength)

		if conf.MaxSrcFileSize > 0 && contentLength > conf.MaxSrcFileSize {
			return ctx, func() {}, errSourceFileTooBig
		}
	}

	buf := downloadBufPool.Get(contentLength)
	cancel := func() {
		downloadBufPool.Put(buf)
	}

	body := res.Body

	if conf.MaxSrcFileSize > 0 {
		body = &limitReader{r: body, left: conf.MaxSrcFileSize}
	}

	imgtype, err := checkTypeAndDimensions(io.TeeReader(body, buf))
	if err != nil {
		return ctx, cancel, err
	}

	if _, err = buf.ReadFrom(body); err != nil {
		return ctx, cancel, newError(404, err.Error(), msgSourceImageIsUnreachable)
	}

	ctx = context.WithValue(ctx, imageTypeCtxKey, imgtype)
	ctx = context.WithValue(ctx, imageDataCtxKey, buf)

	return ctx, cancel, nil
}

func downloadImage(ctx context.Context) (context.Context, context.CancelFunc, error) {
	url := getImageURL(ctx)

	if newRelicEnabled {
		newRelicCancel := startNewRelicSegment(ctx, "Downloading image")
		defer newRelicCancel()
	}

	if prometheusEnabled {
		defer startPrometheusDuration(prometheusDownloadDuration)()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ctx, func() {}, newError(404, err.Error(), msgSourceImageIsUnreachable)
	}

	req.Header.Set("User-Agent", conf.UserAgent)

	res, err := downloadClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return ctx, func() {}, newError(404, err.Error(), msgSourceImageIsUnreachable)
	}

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		msg := fmt.Sprintf("Can't download image; Status: %d; %s", res.StatusCode, string(body))
		return ctx, func() {}, newError(404, msg, msgSourceImageIsUnreachable)
	}

	return readAndCheckImage(ctx, res)
}

func getImageType(ctx context.Context) imageType {
	return ctx.Value(imageTypeCtxKey).(imageType)
}

func getImageData(ctx context.Context) *bytes.Buffer {
	return ctx.Value(imageDataCtxKey).(*bytes.Buffer)
}
