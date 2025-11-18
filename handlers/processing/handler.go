package processing

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	optionsparser "github.com/imgproxy/imgproxy/v3/options/parser"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/workers"
)

// HandlerContext provides access to shared handler dependencies
type HandlerContext interface {
	Workers() *workers.Workers
	ClientFeaturesDetector() *clientfeatures.Detector
	FallbackImage() auximageprovider.Provider
	ImageDataFactory() *imagedata.Factory
	Security() *security.Checker
	OptionsParser() *optionsparser.Parser
	Processor() *processing.Processor
	Cookies() *cookies.Cookies
	Monitoring() *monitoring.Monitoring
	ErrorReporter() *errorreport.Reporter
}

// Handler handles image processing requests
type Handler struct {
	HandlerContext

	stream *stream.Handler // Stream handler for raw image streaming
	config *Config         // Handler configuration
}

// New creates new handler object
func New(
	hCtx HandlerContext,
	stream *stream.Handler,
	config *Config,
) (*Handler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Handler{
		HandlerContext: hCtx,
		config:         config,
		stream:         stream,
	}, nil
}

// Execute handles the image processing request
func (h *Handler) Execute(
	reqID string,
	rw server.ResponseWriter,
	req *http.Request,
) error {
	// Increment the number of requests in progress
	h.Monitoring().Stats().IncRequestsInProgress()
	defer h.Monitoring().Stats().DecRequestsInProgress()

	ctx := req.Context()

	// Verify URL signature and extract image url and processing options
	r, err := h.newRequest(ctx, req)
	if err != nil {
		return err
	}

	// if processing options indicate raw image streaming, stream it and return
	if r.opts.GetBool(keys.Raw, false) {
		return h.stream.Execute(ctx, req, r.imageURL, reqID, r.opts, rw)
	}

	r.reqID = reqID
	r.req = req
	r.rw = rw
	r.config = h.config

	return r.execute(ctx)
}

// newRequest extracts image url and processing options from request URL and verifies them
func (h *Handler) newRequest(
	ctx context.Context,
	req *http.Request,
) (*request, error) {
	// let's extract signature and valid request path from a request
	path, signature, err := handlers.SplitPathSignature(req)
	if err != nil {
		return nil, err
	}

	// verify the signature (if any)
	if err = h.Security().VerifySignature(signature, path); err != nil {
		return nil, errctx.Wrap(err, 0, errctx.WithCategory(handlers.CategorySecurity))
	}

	// parse image url and processing options
	features := h.ClientFeaturesDetector().Features(req.Header)
	o, imageURL, err := h.OptionsParser().ParsePath(path, &features)
	if err != nil {
		return nil, errctx.Wrap(err, 0, errctx.WithCategory(handlers.CategoryPathParsing))
	}

	// get image origin and create monitoring meta object
	imageOrigin := monitoring.MetaURLOrigin(imageURL)

	mm := monitoring.Meta{
		monitoring.MetaSourceImageURL:    imageURL,
		monitoring.MetaSourceImageOrigin: imageOrigin,
		monitoring.MetaOptions:           o.Map(),
	}

	// set error reporting and monitoring context
	errorreport.SetMetadata(req, "Source Image URL", imageURL)
	errorreport.SetMetadata(req, "Source Image Origin", imageOrigin)
	errorreport.SetMetadata(req, "Options", o.NestedMap())

	h.Monitoring().SetMetadata(ctx, mm)

	// verify that image URL came from the valid source
	err = h.Security().VerifySourceURL(imageURL)
	if err != nil {
		return nil, errctx.Wrap(err, 0, errctx.WithCategory(handlers.CategorySecurity))
	}

	return &request{
		HandlerContext: h,

		imageURL:       imageURL,
		path:           path,
		opts:           o,
		features:       &features,
		monitoringMeta: mm,
		req:            req,
	}, nil
}
