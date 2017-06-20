package main

import (
	"bytes"
	"errors"
	"image"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const chunkSize = 4096

func checkTypeAndDimensions(b []byte) error {
	imgconf, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return err
	}
	if imgconf.Width > conf.MaxSrcDimension || imgconf.Height > conf.MaxSrcDimension {
		return errors.New("File is too big")
	}
	return nil
}

func readAndCheckImage(res *http.Response) ([]byte, error) {
	b := make([]byte, chunkSize)
	n, err := res.Body.Read(b)
	if err != nil {
		return nil, err
	}

	if err = checkTypeAndDimensions(b[:n]); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(b[:n])

	if res.ContentLength > 0 {
		buf.Grow(int(res.ContentLength))
	}

	if _, err = buf.ReadFrom(res.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func downloadImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return readAndCheckImage(res)
}
