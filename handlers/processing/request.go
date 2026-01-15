package processing

import (
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// request holds the parameters and state for a single request
type request struct {
	HandlerContext

	reqID          string
	req            *http.Request
	rw             server.ResponseWriter
	config         *Config
	opts           *options.Options
	imageURL       string
	path           string
	monitoringMeta monitoring.Meta
	features       *clientfeatures.Features
}

// execute handles the actual processing logic
func (r *request) execute() *server.Error {
	outFormat := options.Get(r.opts, keys.Format, imagetype.Unknown)

	// Check if we can save the resulting image
	canSave := vips.SupportsSave(outFormat) ||
		outFormat == imagetype.Unknown ||
		outFormat == imagetype.SVG

	if !canSave {
		return server.NewError(handlers.NewCantSaveError(outFormat), handlers.ErrCategoryPathParsing)
	}

	ctx := r.req.Context()

	// Acquire worker
	releaseWorker, err := r.acquireWorker()
	if err != nil {
		return server.NewError(err, handlers.ErrCategoryQueue)
	}
	defer releaseWorker()

	// Deal with processing image counter
	r.Monitoring().Stats().IncImagesInProgress()
	defer r.Monitoring().Stats().DecImagesInProgress()

	// Response status code is OK by default
	statusCode := http.StatusOK

	// Request headers
	imgRequestHeaders := r.makeImageRequestHeaders()

	// create download options
	do, err := r.makeDownloadOptions(imgRequestHeaders)
	if err != nil {
		return server.NewError(err, handlers.ErrCategoryDownload)
	}

	// Fetch image actual
	originData, originHeaders, err := r.fetchImage(do)
	if err == nil {
		defer originData.Close() // if any originData has been opened, we need to close it
	}

	// Check that image detection didn't take too long
	if terr := server.CheckTimeout(ctx); terr != nil {
		return server.NewError(terr, handlers.ErrCategoryTimeout)
	}

	// Respond with NotModified if image was not modified
	var nmErr fetcher.NotModifiedError

	if errors.As(err, &nmErr) {
		r.rw.SetOriginHeaders(nmErr.Headers())

		r.respondWithNotModified()

		return nil
	}

	// Prepare to write image response headers
	r.rw.SetOriginHeaders(originHeaders)

	// If error is not related to NotModified, respond with fallback image and replace image data
	if err != nil {
		originData, statusCode, err = r.handleDownloadError(err)
		if err != nil {
			return server.NewError(err, handlers.ErrCategoryDownload)
		}
	}

	// Check if image supports load from origin format
	if !vips.SupportsLoad(originData.Format()) {
		return server.NewError(
			handlers.NewCantLoadError(ctx, originData.Format()),
			handlers.ErrCategoryPathParsing,
		)
	}

	// Actually process the image
	result, err := r.processImage(originData)

	// Let's close resulting image data only if it differs from the source image data
	if result != nil && result.OutData != nil && result.OutData != originData {
		defer result.OutData.Close()
	}

	// First, check if the processing error wasn't caused by an image data error
	if derr := originData.Error(); derr != nil {
		return server.NewError(r.wrapDownloadingErr(derr), handlers.ErrCategoryDownload)
	}

	// If it wasn't, than it was a processing error
	if err != nil {
		return server.NewError(err, handlers.ErrCategoryProcessing)
	}

	// Write debug headers. It seems unlogical to move they to responsewriter since they're
	// not used anywhere else.
	dhErr := r.writeDebugHeaders(result, originData)
	if dhErr != nil {
		return dhErr
	}

	// Responde with actual image
	return r.respondWithImage(statusCode, result.OutData)
}
