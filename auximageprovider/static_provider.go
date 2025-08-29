package auximageprovider

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

// StaticConfig represents the configuration for the [staticProvider].
type StaticConfig struct {
	ImageKind ImageKind

	Base64Data string
	Path       string
	URL        string
}

// NewDefaultStaticConfig returns a new [StaticConfig] instance with default values.
func NewDefaultStaticConfig(kind ImageKind) *StaticConfig {
	return &StaticConfig{ImageKind: kind}
}

// LoadFromEnv loads provider config variables from environment based on the image kind.
func (c *StaticConfig) LoadFromEnv() *StaticConfig {
	switch c.ImageKind {
	case ImageKindWatermark:
		c.Base64Data = config.WatermarkData
		c.Path = config.WatermarkPath
		c.URL = config.WatermarkURL
	case ImageKindFallback:
		c.Base64Data = config.FallbackImageData
		c.Path = config.FallbackImagePath
		c.URL = config.FallbackImageURL
	}

	return c
}

// staticProvider is a simple implementation of [Provider], which returns
// a static saved image data and headers.
type staticProvider struct {
	data    imagedata.ImageData
	headers http.Header
}

// NewStaticProvider creates a new [staticProvider] instance.
func NewStaticProvider(ctx context.Context, cfg *StaticConfig) (Provider, error) {
	var (
		data    imagedata.ImageData
		headers http.Header
		err     error
	)

	switch {
	case len(cfg.Base64Data) > 0:
		data, err = imagedata.NewFromBase64(cfg.Base64Data)
	case len(cfg.Path) > 0:
		data, err = imagedata.NewFromPath(cfg.Path)
	case len(cfg.URL) > 0:
		data, headers, err = imagedata.DownloadSync(
			ctx, cfg.URL, string(cfg.ImageKind), imagedata.DownloadOptions{},
		)
	default:
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if headers == nil {
		headers = make(http.Header)
	}

	return &staticProvider{
		data:    data,
		headers: headers,
	}, nil
}

// Get returns the static image data and headers.
func (p *staticProvider) Get(
	_ context.Context,
	_ *options.ProcessingOptions,
) (imagedata.ImageData, http.Header, error) {
	return p.data, p.headers, nil
}
