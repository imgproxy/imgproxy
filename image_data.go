package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
)

type imageData struct {
	Data []byte
	Type imageType

	cancel context.CancelFunc
}

func (d *imageData) Close() {
	if d.cancel != nil {
		d.cancel()
	}
}

func getWatermarkData() (*imageData, error) {
	if len(conf.WatermarkData) > 0 {
		return base64ImageData(conf.WatermarkData, "watermark")
	}

	if len(conf.WatermarkPath) > 0 {
		return fileImageData(conf.WatermarkPath, "watermark")
	}

	if len(conf.WatermarkURL) > 0 {
		return remoteImageData(conf.WatermarkURL, "watermark")
	}

	return nil, nil
}

func getFallbackImageData() (*imageData, error) {
	if len(conf.FallbackImageData) > 0 {
		return base64ImageData(conf.FallbackImageData, "fallback image")
	}

	if len(conf.FallbackImagePath) > 0 {
		return fileImageData(conf.FallbackImagePath, "fallback image")
	}

	if len(conf.FallbackImageURL) > 0 {
		return remoteImageData(conf.FallbackImageURL, "fallback image")
	}

	return nil, nil
}

func base64ImageData(encoded, desc string) (*imageData, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s data: %s", desc, err)
	}

	imgtype, err := checkTypeAndDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s: %s", desc, err)
	}

	return &imageData{Data: data, Type: imgtype}, nil
}

func fileImageData(path, desc string) (*imageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	imgdata, err := readAndCheckImage(f, int(fi.Size()))
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	return imgdata, err
}

func remoteImageData(imageURL, desc string) (*imageData, error) {
	res, err := requestImage(imageURL)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("Can't download %s: %s", desc, err)
	}

	imgdata, err := readAndCheckImage(res.Body, int(res.ContentLength))
	if err != nil {
		return nil, fmt.Errorf("Can't download %s: %s", desc, err)
	}

	return imgdata, err
}
