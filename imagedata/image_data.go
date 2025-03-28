package imagedata

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
)

var (
	Watermark     *ImageData
	CWWatermark     *ImageData
	BWWatermark     *ImageData
	BWWatermarkV2     *ImageData
	ArtifactMap map[string] *ImageData
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

func Init() error {
	initRead()

	if err := initDownloading(); err != nil {
		return err
	}

	if err := loadWatermarkAndArtifacts(); err != nil {
		return err
	}

	if err := loadFallbackImage(); err != nil {
		return err
	}

	return nil
}

func loadWatermarkAndArtifacts() error {
	ctx := context.Background()

	watermarks := map[string]**ImageData{
		"s3://m-aeplimages/watermarks/cw_watermark.png": &CWWatermark,
		"s3://m-aeplimages/watermarks/bw_watermark.png" : &BWWatermark,
		"s3://m-aeplimages/watermarks/bw_watermark_v2.png" : &BWWatermarkV2,
	}

	artifacts := map[string]string{
		"1": "s3://m-aeplimages/artifacts/editorial_template_*.png",
		"2": "s3://m-aeplimages/artifacts/editorial_template_bw_*.png",
		"3": "s3://m-aeplimages/artifacts/ios_ad_template_*.png",
		"4": "s3://m-aeplimages/artifacts/android_ad_template_*.png",
		"5": "s3://m-aeplimages/artifacts/bs6_*.png",
		"6": "s3://m-aeplimages/artifacts/bs6_without_tooltip_*.png",
		"7": "s3://m-aeplimages/artifacts/bs6_without_tooltip_v1_*.png",
		"8": "s3://m-aeplimages/artifacts/mobility_template_*.png",
		"9": "s3://m-aeplimages/artifacts/editorial_template_bw_v2_*.png",
	}

	artifactsSizesMap := map[string][]string{
		"1": {"642x336"},
		"2": {"642x336"},
		"3": {"642x361"},
		"4": {"559x314"},
		"5": {"110x61", "160x89", "272x153", "393x221", "476x268", "559x314", "600x337", "642x361", "762x429"},
		"6": {"110x61", "160x89", "272x153", "393x221", "476x268", "559x314", "600x337", "642x361", "762x429"},
		"7": {"110x61", "160x89", "272x153", "393x221", "476x268", "559x314", "600x337", "642x361", "762x429"},
		"8": {"642x336"},
		"9": {"642x336"},
	}

	// Download watermarks
	for url, _ := range watermarks {
		download, err := Download(ctx, url, "watermark", DownloadOptions{}, security.DefaultOptions())
		if err != nil {
			return fmt.Errorf("failed to download watermark from %s: %w", url, err)
		}
		*watermarks[url] = download // Assign the downloaded image data to the pointer
	}

	// Ensure ArtifactMap is initialized
	if ArtifactMap == nil {
		ArtifactMap = make(map[string]*ImageData)
	}

	// Download artifacts
	for artifactType, artifactPath := range artifacts {
		sizes, exists := artifactsSizesMap[artifactType]
		if !exists {
			continue
		}

		for _, size := range sizes {
			artifactURL := strings.Replace(artifactPath, "*", size, 1)
			artifact, err := Download(ctx, artifactURL, "watermark", DownloadOptions{}, security.DefaultOptions())
			if err != nil {
				return fmt.Errorf("failed to download artifact %s (%s): %w", artifactType, size, err)
			}
			ArtifactMap[fmt.Sprintf("%s_%s", artifactType, size)] = artifact
		}
	}

	return nil
}


func loadFallbackImage() (err error) {
	switch {
	case len(config.FallbackImageData) > 0:
		FallbackImage, err = FromBase64(config.FallbackImageData, "fallback image", security.DefaultOptions())
	case len(config.FallbackImagePath) > 0:
		FallbackImage, err = FromFile(config.FallbackImagePath, "fallback image", security.DefaultOptions())
	case len(config.FallbackImageURL) > 0:
		FallbackImage, err = Download(context.Background(), config.FallbackImageURL, "fallback image", DownloadOptions{Header: nil, CookieJar: nil}, security.DefaultOptions())
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

func FromBase64(encoded, desc string, secopts security.Options) (*ImageData, error) {
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))
	size := 4 * (len(encoded)/3 + 1)

	imgdata, err := readAndCheckImage(dec, size, secopts)
	if err != nil {
		return nil, fmt.Errorf("Can't decode %s: %s", desc, err)
	}

	return imgdata, nil
}

func FromFile(path, desc string, secopts security.Options) (*ImageData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	imgdata, err := readAndCheckImage(f, int(fi.Size()), secopts)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s: %s", desc, err)
	}

	return imgdata, nil
}

func Download(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (*ImageData, error) {
	imgdata, err := download(ctx, imageURL, opts, secopts)
	if err != nil {
		return nil, ierrors.Wrap(
			err, 0,
			ierrors.WithPrefix(fmt.Sprintf("Can't download %s", desc)),
		)
	}

	return imgdata, nil
}
