package imagestreamer

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/options"
)

// Service is a struct which stores dependencies for the image streaming service
type Service struct {
	config              *Config
	fetcher             *imagefetcher.Fetcher
	headerWriterFactory *headerwriter.Factory
}

// Request holds parameters for the image streaming options
// NOTE: This struct looks like it could be reused in the processing handler as a request context
type Request struct {
	UserRequest       *http.Request              // Original user request to imgproxy
	ImageURL          string                     // URL of the image to be streamed
	ReqID             string                     // Unique identifier for the request
	ProcessingOptions *options.ProcessingOptions // Processing options for the image
	Rw                http.ResponseWriter        // Response writer to write the streamed image
}

// NewService creates a new service instance with the provided configuration
func NewService(config *Config, fetcher *imagefetcher.Fetcher, headerWriterFactory *headerwriter.Factory) *Service {
	return &Service{config: config, fetcher: fetcher, headerWriterFactory: headerWriterFactory}
}

// Stream streams the image based on the provided request
func (f *Service) Stream(ctx context.Context, r *Request) {
	s := &streamer{
		fetcher:             f.fetcher,
		headerWriterFactory: f.headerWriterFactory,
		config:              f.config,
		p:                   r,
	}

	s.Stream(ctx)
}
