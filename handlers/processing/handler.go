package processing

import (
	"context"
	"net/http"
	"net/url"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/semaphores"
)

// Handler handles image processing requests
type Handler struct {
	hw             *headerwriter.Writer // Configured HeaderWriter instance
	stream         *stream.Handler      // Stream handler for raw image streaming
	config         *Config              // Handler configuration
	semaphores     *semaphores.Semaphores
	fallbackImage  auximageprovider.Provider
	watermarkImage auximageprovider.Provider
	idf            *imagedata.Factory
}

// New creates new handler object
func New(
	stream *stream.Handler,
	hw *headerwriter.Writer,
	semaphores *semaphores.Semaphores,
	fi auximageprovider.Provider,
	wi auximageprovider.Provider,
	idf *imagedata.Factory,
	config *Config,
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
		idf:            idf,
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

	// Verify URL signature and extract image url and processing options
	imageURL, po, mm, err := h.newRequest(ctx, imageRequest)
	if err != nil {
		return err
	}

	// if processing options indicate raw image streaming, stream it and return
	if po.Raw {
		return h.stream.Execute(ctx, imageRequest, imageURL, reqID, po, rw)
	}

	req := &request{
		handler:        h,
		imageRequest:   imageRequest,
		reqID:          reqID,
		rw:             rw,
		config:         h.config,
		po:             po,
		imageURL:       imageURL,
		monitoringMeta: mm,
		semaphores:     h.semaphores,
		hwr:            h.hw.NewRequest(),
		idf:            h.idf,
	}

	return req.execute(ctx)
}

// newRequest extracts image url and processing options from request URL and verifies them
func (h *Handler) newRequest(
	ctx context.Context,
	imageRequest *http.Request,
) (string, *options.ProcessingOptions, monitoring.Meta, error) {
	// let's extract signature and valid request path from a request
	path, signature, err := splitPathSignature(imageRequest, h.config)
	if err != nil {
		return "", nil, nil, err
	}

	// verify the signature (if any)
	if err = security.VerifySignature(signature, path); err != nil {
		return "", nil, nil, ierrors.Wrap(err, 0, ierrors.WithCategory(categorySecurity))
	}

	// parse image url and processing options
	po, imageURL, err := options.ParsePath(path, imageRequest.Header)
	if err != nil {
		return "", nil, nil, ierrors.Wrap(err, 0, ierrors.WithCategory(categoryPathParsing))
	}

	// get image origin and create monitoring meta object
	imageOrigin := imageOrigin(imageURL)

	mm := monitoring.Meta{
		monitoring.MetaSourceImageURL:    imageURL,
		monitoring.MetaSourceImageOrigin: imageOrigin,
		monitoring.MetaProcessingOptions: po.Diff().Flatten(),
	}

	// set error reporting and monitoring context
	errorreport.SetMetadata(imageRequest, "Source Image URL", imageURL)
	errorreport.SetMetadata(imageRequest, "Source Image Origin", imageOrigin)
	errorreport.SetMetadata(imageRequest, "Processing Options", po)

	monitoring.SetMetadata(ctx, mm)

	// verify that image URL came from the valid source
	err = security.VerifySourceURL(imageURL)
	if err != nil {
		return "", nil, mm, ierrors.Wrap(err, 0, ierrors.WithCategory(categorySecurity))
	}

	return imageURL, po, mm, nil
}

// imageOrigin extracts image origin from URL
func imageOrigin(imageURL string) string {
	if u, uerr := url.Parse(imageURL); uerr == nil {
		return u.Scheme + "://" + u.Host
	}

	return ""
}
