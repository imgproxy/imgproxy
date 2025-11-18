package processing

import (
	"context"
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/clientfeatures"
	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// request holds the parameters and state for a single request request
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
func (r *request) execute(ctx context.Context) error {
	outFormat := options.Get(r.opts, keys.Format, imagetype.Unknown)

	// Check if we can save the resulting image
	canSave := vips.SupportsSave(outFormat) ||
		outFormat == imagetype.Unknown ||
		outFormat == imagetype.SVG

	if !canSave {
		return handlers.NewCantSaveError(outFormat)
	}

	// Acquire worker
	releaseWorker, err := r.acquireWorker(ctx)
	if err != nil {
		return err
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
	do, err := r.makeDownloadOptions(ctx, imgRequestHeaders)
	if err != nil {
		return err
	}

	// Fetch image actual
	originData, originHeaders, err := r.fetchImage(ctx, do)
	if err == nil {
		defer originData.Close() // if any originData has been opened, we need to close it
	}

	// Check that image detection didn't take too long
	if terr := server.CheckTimeout(ctx); terr != nil {
		return errctx.Wrap(terr, 0, errctx.WithCategory(handlers.CategoryTimeout))
	}

	// Respond with NotModified if image was not modified
	var nmErr fetcher.NotModifiedError

	if errors.As(err, &nmErr) {
		r.rw.SetOriginHeaders(nmErr.Headers())

		return r.respondWithNotModified()
	}

	// Prepare to write image response headers
	r.rw.SetOriginHeaders(originHeaders)

	// If error is not related to NotModified, respond with fallback image and replace image data
	if err != nil {
		originData, statusCode, err = r.handleDownloadError(ctx, err)
		if err != nil {
			return err
		}
	}

	// Check if image supports load from origin format
	if !vips.SupportsLoad(originData.Format()) {
		return handlers.NewCantLoadError(originData.Format())
	}

	// Actually process the image
	result, err := r.processImage(ctx, originData)

	// Let's close resulting image data only if it differs from the source image data
	if result != nil && result.OutData != nil && result.OutData != originData {
		defer result.OutData.Close()
	}

	// First, check if the processing error wasn't caused by an image data error
	if derr := originData.Error(); derr != nil {
		return r.wrapDownloadingErr(derr)
	}

	// If it wasn't, than it was a processing error
	if err != nil {
		return errctx.Wrap(err, 0, errctx.WithCategory(handlers.CategoryProcessing))
	}

	// Write debug headers. It seems unlogical to move they to responsewriter since they're
	// not used anywhere else.
	err = r.writeDebugHeaders(result, originData)
	if err != nil {
		return err
	}

	// Responde with actual image
	err = r.respondWithImage(statusCode, result.OutData)
	if err != nil {
		return err
	}

	return nil
}
