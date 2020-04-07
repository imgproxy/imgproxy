package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
)

func getFallbackData() (*imageData, error) {
	if len(conf.FallbackData) > 0 {
		return base64FallbackData(conf.FallbackData)
	}

	if len(conf.FallbackPath) > 0 {
		return fileFallbackData(conf.FallbackPath)
	}

	if len(conf.FallbackURL) > 0 {
		return remoteFallbackData(conf.FallbackURL)
	}

	return nil, nil
}

func base64FallbackData(encoded string) (*imageData, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("Can't decode fallback data: %s", err)
	}

	imgtype, err := checkTypeAndDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Can't decode fallback: %s", err)
	}

	return &imageData{Data: data, Type: imgtype}, nil
}

func fileFallbackData(path string) (*imageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Can't read fallback: %s", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("Can't read fallback: %s", err)
	}

	imgdata, err := readAndCheckImage(f, int(fi.Size()))
	if err != nil {
		return nil, fmt.Errorf("Can't read fallback: %s", err)
	}

	return imgdata, err
}

func remoteFallbackData(imageURL string) (*imageData, error) {
	res, err := requestImage(imageURL)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Can't download fallback: %s", err)
	}

	imgdata, err := readAndCheckImage(res.Body, int(res.ContentLength))
	if err != nil {
		return nil, fmt.Errorf("Can't download fallback: %s", err)
	}

	return imgdata, err
}
