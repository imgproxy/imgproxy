package main

import (
	"bufio"
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
)

var downloadBufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type netReader struct {
	reader *bufio.Reader
	buf    *bytes.Buffer
}

func newNetReader(r io.Reader, buf *bytes.Buffer) *netReader {
	return &netReader{
		reader: bufio.NewReader(r),
		buf:    buf,
	}
}

func (r *netReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if err == nil {
		r.buf.Write(p[:n])
	}
	return
}

func (r *netReader) Peek(n int) ([]byte, error) {
	return r.reader.Peek(n)
}

func (r *netReader) ReadAll() error {
	if _, err := r.buf.ReadFrom(r.reader); err != nil {
		return err
	}
	return nil
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

	downloadClient = &http.Client{
		Timeout:   time.Duration(conf.DownloadTimeout) * time.Second,
		Transport: transport,
	}
}

func checkTypeAndDimensions(r io.Reader) (imageType, error) {
	imgconf, imgtypeStr, err := image.DecodeConfig(r)
	imgtype, imgtypeOk := imageTypes[imgtypeStr]

	if err != nil {
		return imageTypeUnknown, err
	}
	if imgconf.Width > conf.MaxSrcDimension || imgconf.Height > conf.MaxSrcDimension {
		return imageTypeUnknown, errors.New("Source image is too big")
	}
	if imgconf.Width*imgconf.Height > conf.MaxSrcResolution {
		return imageTypeUnknown, errors.New("Source image is too big")
	}
	if !imgtypeOk || !vipsTypeSupportLoad[imgtype] {
		return imageTypeUnknown, errors.New("Source image type not supported")
	}

	return imgtype, nil
}

func readAndCheckImage(ctx context.Context, res *http.Response) (context.Context, context.CancelFunc, error) {
	buf := downloadBufPool.Get().(*bytes.Buffer)
	cancel := func() {
		buf.Reset()
		downloadBufPool.Put(buf)
	}

	nr := newNetReader(res.Body, buf)

	imgtype, err := checkTypeAndDimensions(nr)
	if err != nil {
		return ctx, cancel, err
	}

	if err = nr.ReadAll(); err == nil {
		ctx = context.WithValue(ctx, imageTypeCtxKey, imgtype)
		ctx = context.WithValue(ctx, imageDataCtxKey, nr.buf)
	}

	return ctx, cancel, err
}

func downloadImage(ctx context.Context) (context.Context, context.CancelFunc, error) {
	url := fmt.Sprintf("%s%s", conf.BaseURL, getImageURL(ctx))

	res, err := downloadClient.Get(url)
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
