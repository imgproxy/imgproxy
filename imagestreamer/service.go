package imagestreamer

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
)

// Service is a struct which stores dependencies for the image streaming service
type Service struct {
	config              *Config
	fetcher             *imagefetcher.Fetcher
	headerWriterFactory *headerwriter.Factory
}

// NewService creates a new service instance with the provided configuration
func NewService(config *Config, fetcher *imagefetcher.Fetcher, headerWriterFactory *headerwriter.Factory) *Service {
	return &Service{config: config, fetcher: fetcher, headerWriterFactory: headerWriterFactory}
}

// Stream streams the image based on the provided request
func (f *Service) Stream(ctx context.Context, r *Request, rr http.ResponseWriter) {
	s := &streamer{
		fetcher:             f.fetcher,
		headerWriterFactory: f.headerWriterFactory,
		config:              f.config,
		p:                   r,
		rw:                  rr,
	}

	s.Stream(ctx)
}
