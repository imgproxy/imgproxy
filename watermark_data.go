package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

type watermarkData struct {
	data    []byte
	imgtype imageType
}

func getWatermarkData() (*watermarkData, error) {
	if len(conf.WatermarkData) > 0 {
		data, imgtype, err := base64WatermarkData()

		if err != nil {
			return nil, err
		}

		return &watermarkData{data, imgtype}, err
	}

	if len(conf.WatermarkPath) > 0 {
		data, imgtype, err := fileWatermarkData()

		if err != nil {
			return nil, err
		}

		return &watermarkData{data, imgtype}, err
	}

	if len(conf.WatermarkURL) > 0 {
		b, imgtype, cancel, err := remoteWatermarkData()
		defer cancel()

		if err != nil {
			return nil, err
		}

		data := make([]byte, len(b))
		copy(data, b)

		return &watermarkData{data, imgtype}, err
	}

	return nil, nil
}

func base64WatermarkData() ([]byte, imageType, error) {
	data, err := base64.StdEncoding.DecodeString(conf.WatermarkData)
	if err != nil {
		return nil, imageTypeUnknown, fmt.Errorf("Can't decode watermark data: %s", err)
	}

	imgtype, err := checkTypeAndDimensions(bytes.NewReader(data))
	if err != nil {
		return nil, imageTypeUnknown, fmt.Errorf("Can't decode watermark: %s", err)
	}

	return data, imgtype, nil
}

func fileWatermarkData() ([]byte, imageType, error) {
	f, err := os.Open(conf.WatermarkPath)
	if err != nil {
		return nil, imageTypeUnknown, fmt.Errorf("Can't read watermark: %s", err)
	}

	imgtype, err := checkTypeAndDimensions(f)
	if err != nil {
		return nil, imageTypeUnknown, fmt.Errorf("Can't decode watermark: %s", err)
	}

	// Return to the beginning of the file
	f.Seek(0, 0)

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, imageTypeUnknown, fmt.Errorf("Can't read watermark: %s", err)
	}

	return data, imgtype, nil
}

func remoteWatermarkData() ([]byte, imageType, context.CancelFunc, error) {
	ctx := context.WithValue(context.Background(), imageURLCtxKey, conf.WatermarkURL)
	ctx, cancel, err := downloadImage(ctx)

	if err != nil {
		return nil, imageTypeUnknown, cancel, fmt.Errorf("Can't download watermark: %s", err)
	}

	return getImageData(ctx).Bytes(), getImageType(ctx), cancel, err
}
