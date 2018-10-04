package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

var downloadClient *http.Client

type netReader struct {
	reader *bufio.Reader
	buf    *bytes.Buffer
}

func newNetReader(r io.Reader) *netReader {
	return &netReader{
		reader: bufio.NewReader(r),
		buf:    bytes.NewBuffer([]byte{}),
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

func (r *netReader) ReadAll() ([]byte, error) {
	if _, err := r.buf.ReadFrom(r.reader); err != nil {
		return []byte{}, err
	}
	return r.buf.Bytes(), nil
}

func (r *netReader) GrowBuf(s int) {
	r.buf.Grow(s)
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

func readAndCheckImage(res *http.Response) ([]byte, imageType, error) {
	nr := newNetReader(res.Body)

	imgtype, err := checkTypeAndDimensions(nr)
	if err != nil {
		return nil, imageTypeUnknown, err
	}

	if res.ContentLength > 0 {
		nr.GrowBuf(int(res.ContentLength))
	}

	b, err := nr.ReadAll()

	return b, imgtype, err
}

func downloadImage(url string) ([]byte, imageType, error) {
	fullURL := fmt.Sprintf("%s%s", conf.BaseURL, url)

	res, err := downloadClient.Get(fullURL)
	if err != nil {
		return nil, imageTypeUnknown, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, imageTypeUnknown, fmt.Errorf("Can't download image; Status: %d; %s", res.StatusCode, string(body))
	}

	return readAndCheckImage(res)
}
