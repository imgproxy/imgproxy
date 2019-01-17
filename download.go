package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"io"
	"io/ioutil"
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
	errSourceResolutionTooBig      = newError(422, "Source image resolution are too big", "Invalid source image")
	errSourceImageTypeNotSupported = newError(422, "Source image type not supported", "Invalid source image")
)

const msgSourceImageIsUnreachable = "Source image is unreachable"

var downloadBufPool *bufPool

func initDownloading() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	if conf.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if conf.LocalFileSystemRoot != "" {
		transport.RegisterProtocol("local", http.NewFileTransport(http.Dir(conf.LocalFileSystemRoot)))
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

	downloadBufPool = newBufPool(conf.Concurrency, conf.DownloadBufferSize)
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
	if err != nil {
		return imageTypeUnknown, errSourceImageTypeNotSupported
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
	buf := downloadBufPool.get()
	cancel := func() {
		downloadBufPool.put(buf)
	}

	imgtype, err := checkTypeAndDimensions(io.TeeReader(res.Body, buf))
	if err != nil {
		return ctx, cancel, err
	}

	if _, err = buf.ReadFrom(res.Body); err != nil {
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
	if err != nil {
		return ctx, func() {}, newError(404, err.Error(), msgSourceImageIsUnreachable)
	}
	defer res.Body.Close()

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
