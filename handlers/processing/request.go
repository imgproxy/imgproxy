package processing

import (
	"context"
	"errors"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// request holds the parameters and state for a single processing request execution
type request = handlers.Request[*Handler, *options.ProcessingOptions]

// execute handles the actual processing logic
func execute(ctx context.Context, r *request) error {
	// Check if we can save the resulting image
	canSave := vips.SupportsSave(r.Options.Format) ||
		r.Options.Format == imagetype.Unknown ||
		r.Options.Format == imagetype.SVG

	if !canSave {
		return handlers.NewCantSaveError(r.Options.Format)
	}

	// Acquire processing semaphore
	releaseProcessingSem, err := r.AcquireProcessingSem(ctx)
	if err != nil {
		return err
	}
	defer releaseProcessingSem()

	// Deal with processing image counter
	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	// Response status code is OK by default
	statusCode := http.StatusOK

	// Request headers
	imgRequestHeaders := r.MakeImageRequestHeaders()

	// create download options
	do := r.MakeDownloadOptions(imgRequestHeaders, r.Options.SecurityOptions)

	// Fetch image actual
	originData, originHeaders, err := r.FetchImage(ctx, do)
	if err == nil {
		defer originData.Close() // if any originData has been opened, we need to close it
	}

	// Check that image detection didn't take too long
	if terr := server.CheckTimeout(ctx); terr != nil {
		return ierrors.Wrap(terr, 0, ierrors.WithCategory(handlers.CategoryTimeout))
	}

	// Respond with NotModified if image was not modified
	var nmErr fetcher.NotModifiedError

	if errors.As(err, &nmErr) {
		r.HeaderWriter.SetOriginHeaders(nmErr.Headers())

		return respondWithNotModified(r)
	}

	// Prepare to write image response headers
	r.HeaderWriter.SetOriginHeaders(originHeaders)

	// If error is not related to NotModified, respond with fallback image and replace image data
	if err != nil {
		originData, statusCode, err = handleDownloadError(ctx, r, err)
		if err != nil {
			return err
		}
	}

	// Check if image supports load from origin format
	if !vips.SupportsLoad(originData.Format()) {
		return handlers.NewCantLoadError(originData.Format())
	}

	// Actually process the image
	result, err := processImage(ctx, r, originData)

	// Let's close resulting image data only if it differs from the source image data
	if result != nil && result.OutData != nil && result.OutData != originData {
		defer result.OutData.Close()
	}

	// First, check if the processing error wasn't caused by an image data error
	if derr := originData.Error(); derr != nil {
		return ierrors.Wrap(derr, 0, ierrors.WithCategory(handlers.CategoryDownload))
	}

	// If it wasn't, than it was a processing error
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryProcessing))
	}

	// Write debug headers. It seems unlogical to move they to headerwriter since they're
	// not used anywhere else.
	err = writeDebugHeaders(r, result, originData)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryImageDataSize))
	}

	// Responde with actual image
	err = respondWithImage(r, statusCode, result.OutData)
	if err != nil {
		return err
	}

	return nil
}
