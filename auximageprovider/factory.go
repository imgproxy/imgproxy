package auximageprovider

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedatanew"
	"github.com/imgproxy/imgproxy/v3/imagedownloader"
	"github.com/imgproxy/imgproxy/v3/security"
)

// Factory is a struct that provides methods to create ImageProvider instances.
type Factory struct {
	downloader *imagedownloader.Downloader
}

// NewFactory creates a new Factory instance with the given downloader.
func NewFactory(downloader *imagedownloader.Downloader) *Factory {
	return &Factory{
		downloader: downloader,
	}
}

// NewMemoryFromBase64 creates a new ImageProvider from a base64 encoded string.
// It stores image in memory.
func (f *Factory) NewMemoryFromBase64(b64 string) (AuxImageProvider, error) {
	img, err := imagedatanew.NewFromBase64(b64, make(http.Header), security.DefaultOptions())
	if err != nil {
		return nil, err
	}

	return newStaticAuxImageProvider(img, make(http.Header)), nil
}

// NewMemoryFromFile creates a new ImageProvider from a local file path.
// It stores image in memory.
func (f *Factory) NewMemoryFromFile(path string) (AuxImageProvider, error) {
	img, err := imagedatanew.NewFromFile(path, make(http.Header), security.DefaultOptions())
	if err != nil {
		return nil, err
	}

	return newStaticAuxImageProvider(img, make(http.Header)), nil
}

// NewMemoryURL creates a new ImageProvider from a URL.
func (f *Factory) NewMemoryURL(ctx context.Context, url string) (AuxImageProvider, error) {
	img, err := f.downloader.DownloadWithDesc(ctx, url, "ImageProvider", imagedownloader.DownloadOptions{}, security.DefaultOptions())
	if err != nil {
		return nil, err
	}

	//nolint:staticcheck
	return newStaticAuxImageProvider(img, img.Headers()), nil
}

// NewMemoryTriple creates a new ImageProvider from either a base64 string, file path, or URL
func (f *Factory) NewMemoryTriple(b64 string, path string, url string) (AuxImageProvider, error) {
	switch {
	case len(b64) > 0:
		return f.NewMemoryFromBase64(b64)
	case len(path) > 0:
		return f.NewMemoryFromFile(path)
	case len(url) > 0:
		return f.NewMemoryURL(context.Background(), url)
	}

	return nil, nil
}
