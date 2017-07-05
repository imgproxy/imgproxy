package main

import (
	"bufio"
	"bytes"
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
)

var downloadClient = http.Client{
	Timeout: time.Duration(conf.DownloadTimeout) * time.Second,
}

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

func checkTypeAndDimensions(r io.Reader) error {
	imgconf, _, err := image.DecodeConfig(r)
	if err != nil {
		return err
	}
	if imgconf.Width > conf.MaxSrcDimension || imgconf.Height > conf.MaxSrcDimension {
		return errors.New("File is too big")
	}
	return nil
}

func readAndCheckImage(res *http.Response) ([]byte, error) {
	nr := newNetReader(res.Body)

	if err := checkTypeAndDimensions(nr); err != nil {
		return nil, err
	}

	if res.ContentLength > 0 {
		nr.GrowBuf(int(res.ContentLength))
	}

	return nr.ReadAll()
}

func downloadImage(url string) ([]byte, error) {
	res, err := downloadClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("Can't download image; Status: %d; %s", res.StatusCode, string(body))
	}

	return readAndCheckImage(res)
}
