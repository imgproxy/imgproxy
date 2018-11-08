package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

var (
	downloadClient  *http.Client
	imageTypeCtxKey = ctxKey("imageType")
	imageDataCtxKey = ctxKey("imageData")

	errSourceDimensionsTooBig      = errors.New("Source image dimensions are too big")
	errSourceResolutionTooBig      = errors.New("Source image resolution are too big")
	errSourceImageTypeNotSupported = errors.New("Source image type not supported")
)

var downloadBufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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
}

func checkDimensions(width, height int) error {
	if width > conf.MaxSrcDimension || height > conf.MaxSrcDimension {
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
	buf := downloadBufPool.Get().(*bytes.Buffer)
	cancel := func() {
		buf.Reset()
		downloadBufPool.Put(buf)
	}

	imgtype, err := checkTypeAndDimensions(io.TeeReader(res.Body, buf))
	if err != nil {
		return ctx, cancel, err
	}

	if _, err = buf.ReadFrom(res.Body); err == nil {
		ctx = context.WithValue(ctx, imageTypeCtxKey, imgtype)
		ctx = context.WithValue(ctx, imageDataCtxKey, buf)
	}

	return ctx, cancel, err
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
		return ctx, func() {}, err
	}

	req.Header.Set("User-Agent", conf.UserAgent)

	res, err := downloadClient.Do(req)
	if err != nil {
		return ctx, func() {}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return ctx, func() {}, fmt.Errorf("Can't download image; Status: %d; %s", res.StatusCode, string(body))
	}

	return readAndCheckImage(ctx, res)
}

func getImageType(ctx context.Context) imageType {
	return ctx.Value(imageTypeCtxKey).(imageType)
}

func getImageData(ctx context.Context) *bytes.Buffer {
	return ctx.Value(imageDataCtxKey).(*bytes.Buffer)
}
