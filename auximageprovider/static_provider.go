package auximageprovider

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

// staticProvider is a simple implementation of ImageProvider, which returns
// a static saved image data and headers.
type staticProvider struct {
	data    imagedata.ImageData
	headers http.Header
}

// Get returns the static image data and headers stored in the provider.
func (s *staticProvider) Get(_ context.Context, po *options.ProcessingOptions) (imagedata.ImageData, http.Header, error) {
	return s.data, s.headers.Clone(), nil
}

// NewStaticFromTriple creates a new ImageProvider from either a base64 string, file path, or URL
func NewStaticProvider(ctx context.Context, c *StaticConfig, desc string) (Provider, error) {
	var (
		data    imagedata.ImageData
		headers = make(http.Header)
		err     error
	)

	switch {
	case len(c.Base64Data) > 0:
		data, err = imagedata.NewFromBase64(c.Base64Data)
	case len(c.Path) > 0:
		data, err = imagedata.NewFromPath(c.Path)
	case len(c.URL) > 0:
		data, headers, err = imagedata.DownloadSync(
			ctx, c.URL, desc, imagedata.DownloadOptions{},
		)
	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &staticProvider{
		data:    data,
		headers: headers,
	}, nil
}
