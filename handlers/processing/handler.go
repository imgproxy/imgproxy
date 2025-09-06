package processing

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/semaphores"
)

// Handler handles image processing requests
type Handler struct {
	config         *handlers.Config     // Handler configuration
	hw             *headerwriter.Writer // Configured HeaderWriter instance
	stream         *stream.Handler      // Stream handler for raw image streaming
	semaphores     *semaphores.Semaphores
	fallbackImage  auximageprovider.Provider
	watermarkImage auximageprovider.Provider
	imageData      *imagedata.Factory
}

// New creates new handler object
func New(
	stream *stream.Handler,
	hw *headerwriter.Writer,
	semaphores *semaphores.Semaphores,
	fi auximageprovider.Provider,
	wi auximageprovider.Provider,
	idf *imagedata.Factory,
	config *handlers.Config,
) (*Handler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Handler{
		hw:             hw,
		config:         config,
		stream:         stream,
		semaphores:     semaphores,
		fallbackImage:  fi,
		watermarkImage: wi,
		imageData:      idf,
	}, nil
}

func (h *Handler) HeaderWriter() *headerwriter.Writer {
	return h.hw
}

func (h *Handler) Semaphores() *semaphores.Semaphores {
	return h.semaphores
}

func (h *Handler) ImageData() *imagedata.Factory {
	return h.imageData
}

func (h *Handler) ParsePath(
	path string,
	headers http.Header,
) (*options.ProcessingOptions, string, error) {
	return options.ParsePath(path, headers)
}

func (h *Handler) SetMonitoringMeta(
	ctx context.Context,
	imageURL, imageOrigin string,
	po *options.ProcessingOptions,
) monitoring.Meta {
	mm := monitoring.Meta{
		monitoring.MetaSourceImageURL:    imageURL,
		monitoring.MetaSourceImageOrigin: imageOrigin,
		monitoring.MetaProcessingOptions: po.Diff().Flatten(),
	}
	monitoring.SetMetadata(ctx, mm)

	// set error reporting and monitoring context
	errorreport.SetMetadata(ctx, "Source Image URL", imageURL)
	errorreport.SetMetadata(ctx, "Source Image Origin", imageOrigin)
	errorreport.SetMetadata(ctx, "Processing Options", po)

	return mm
}

// Execute handles the image processing request
func (h *Handler) Execute(
	reqID string,
	rw http.ResponseWriter,
	req *http.Request,
) error {
	// Increment the number of requests in progress
	stats.IncRequestsInProgress()
	defer stats.DecRequestsInProgress()

	handlerReq, err := handlers.NewRequest(h, h.config, reqID, req, rw)
	if err != nil {
		return err
	}

	ctx := req.Context()

	// if processing options indicate raw image streaming, stream it and return
	if handlerReq.Options.Raw {
		return h.stream.Execute(ctx, req, handlerReq.ImageURL, reqID, handlerReq.Options, rw)
	}

	return execute(ctx, handlerReq)
}
