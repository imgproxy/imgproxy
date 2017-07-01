package main

import (
	"bytes"
	"errors"
	"image"
	"io"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type netReader struct {
	reader io.Reader
	buf    *bytes.Buffer
}

func newNetReader(r io.Reader) *netReader {
	return &netReader{
		reader: r,
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
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return readAndCheckImage(res)
}
