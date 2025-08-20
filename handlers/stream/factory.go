package stream

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
)

// Factory is a struct which stores dependencies for the image streaming service
// NOTE: Probably, we'll use the same factory for all handlers in the future.
type Factory struct {
	config   *Config
	hwConfig *headerwriter.Config
	fetcher  *imagefetcher.Fetcher
}

// New creates a new handler instance with the provided configuration
func New(config *Config, hwConfig *headerwriter.Config, fetcher *imagefetcher.Fetcher) *Factory {
	return &Factory{config: config, fetcher: fetcher, hwConfig: hwConfig}
}

// Stream streams the image based on the provided request
func (h *Factory) NewHandler(ctx context.Context, p *StreamingParams, rr http.ResponseWriter) *Handler {
	return &Handler{
		fetcher:  h.fetcher,
		config:   h.config,
		hwConfig: h.hwConfig,
		params:   p,
		res:      rr,
	}
}
