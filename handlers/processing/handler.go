package processing

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
)

// Handler handles image processing requests
type Handler struct {
	hCtx   handlers.Context // Input context interface
	stream *stream.Handler  // Stream handler for raw image streaming
	config *handlers.Config // Handler configuration
}

type request struct {
	*handlers.Request
	Options *options.ProcessingOptions // Processing options extracted from URL
}

// New creates new handler object
func New(
	context handlers.Context,
	stream *stream.Handler,
	config *handlers.Config,
) (*Handler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Handler{
		hCtx:   context,
		config: config,
		stream: stream,
	}, nil
}

// Execute handles the image processing request
func (h *Handler) Execute(
	reqID string,
	rw http.ResponseWriter,
	imageRequest *http.Request,
) error {
	// Increment the number of requests in progress
	stats.IncRequestsInProgress()
	defer stats.DecRequestsInProgress()

	ctx := imageRequest.Context()

	r, po, err := handlers.NewRequest(h.hCtx, h, imageRequest, h.config, reqID, rw)
	if err != nil {
		return err
	}

	// if processing options indicate raw image streaming, stream it and return
	if po.Raw {
		return h.stream.Execute(ctx, imageRequest, r.ImageURL, reqID, po, rw)
	}

	req := &request{
		Request: r,
		Options: po,
	}

	return execute(ctx, req)
}

func (h *Handler) ParsePath(path string, headers http.Header) (*options.ProcessingOptions, string, error) {
	return options.ParsePath(path, headers)
}

func (h *Handler) CreateMeta(ctx context.Context, imageURL string, po *options.ProcessingOptions) monitoring.Meta {
	imageOrigin := imageOrigin(imageURL)

	mm := monitoring.Meta{
		monitoring.MetaSourceImageURL:    imageURL,
		monitoring.MetaSourceImageOrigin: imageOrigin,
		monitoring.MetaProcessingOptions: po.Diff().Flatten(),
	}

	monitoring.SetMetadata(ctx, mm)

	// NOTE: errorreport needs to be patched (just not in the context of this PR)
	// set error reporting and monitoring context
	// errorreport.SetMetadata(ctx, "Source Image URL", imageURL)
	// errorreport.SetMetadata(ctx, "Source Image Origin", imageOrigin)
	// errorreport.SetMetadata(ctx, "Processing Options", po)

	return mm
}
