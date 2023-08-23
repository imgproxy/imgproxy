package imagedata

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	path2 "path"
	"strings"
	"sync"

	"golang.org/x/net/proxy"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var (
	Watermark     *ImageData
	FallbackImage *ImageData
)

type ImageData struct {
	Type    imagetype.Type
	Data    []byte
	Headers map[string]string

	cancel     context.CancelFunc
	cancelOnce sync.Once
}

func (d *ImageData) Close() {
	d.cancelOnce.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})
}

func (d *ImageData) SetCancel(cancel context.CancelFunc) {
	d.cancel = cancel
}

func Init(dialer proxy.Dialer) error {
	initRead()

	if err := initDownloading(dialer); err != nil {
		return err
	}

	if err := loadWatermark(); err != nil {
		return err
	}

	if err := loadFallbackImage(); err != nil {
		return err
	}

	return nil
}

func loadWatermark() (err error) {
	if len(config.WatermarkData) > 0 {
		Watermark, err = FromBase64(config.WatermarkData, "watermark")
		return
	}

	if len(config.WatermarkPath) > 0 {
		Watermark, err = FromFile(config.WatermarkPath, "watermark")
		return
	}

	if len(config.WatermarkURL) > 0 {
		Watermark, err = Download(config.WatermarkURL, "watermark", nil, nil)
		return
	}

	return nil
}

func loadFallbackImage() (err error) {
	switch {
	case len(config.FallbackImageData) > 0:
		FallbackImage, err = FromBase64(config.FallbackImageData, "fallback image")
	case len(config.FallbackImagePath) > 0:
		FallbackImage, err = FromFile(config.FallbackImagePath, "fallback image")
	case len(config.FallbackImageURL) > 0:
		FallbackImage, err = Download(config.FallbackImageURL, "fallback image", nil, nil)
	default:
		FallbackImage, err = nil, nil
	}

	if FallbackImage != nil && err == nil && config.FallbackImageTTL > 0 {
		if FallbackImage.Headers == nil {
			FallbackImage.Headers = make(map[string]string)
		}
		FallbackImage.Headers["Fallback-Image"] = "1"
	}

	return err
}

func FromBase64(encoded, desc string) (*ImageData, error) {
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
	size := 4 * (len(encoded)/3 + 1)

	imgdata, err := readAndCheckImage(dec, size)
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s: %s", desc, err)
	}

	return imgdata, nil
}

func FromFile(path, desc string) (*ImageData, error) {
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

	return imgdata, nil
}

func Download(imageURL, desc string, header http.Header, jar *cookiejar.Jar) (*ImageData, error) {
	path := path2.Join(config.LocalFileSystemCache, base64.URLEncoding.EncodeToString([]byte(imageURL)))
	if config.LocalFileSystemCache != "" && header.Get("Cache-Control") != "no-cache" {
		imgdata, err := FromFile(path, desc)
		if err == nil {
			fmt.Println("hit")
			return imgdata, nil
		}
	}

	imgdata, err := download(imageURL, header, jar)
	if err != nil {
		if nmErr, ok := err.(*ErrorNotModified); ok {
			nmErr.Message = fmt.Sprintf("Can't download %s: %s", desc, nmErr.Message)
			return nil, nmErr
		}
		return nil, ierrors.WrapWithPrefix(err, 1, fmt.Sprintf("Can't download %s", desc))
	}

	return imgdata, nil
}
