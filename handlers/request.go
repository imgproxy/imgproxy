package handlers

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/structdiff"
)

// Options is an object of URL options extracted from the URL
type Options = structdiff.Diffable

// PathPaser is an interface for URL path parser: it extracts processing options and image path
type Constructor[O Options] interface {
	ParsePath(path string, headers http.Header) (O, string, error)
	CreateMeta(ctx context.Context, imageURL string, po O) monitoring.Meta
}

type Request struct {
	Context        Context               // Input context interface
	Config         *Config               // Handler configuration
	ID             string                // Request ID
	Req            *http.Request         // Original HTTP request
	ResponseWriter http.ResponseWriter   // HTTP response writer
	HeaderWriter   *headerwriter.Request // Header writer request
	ImageURL       string                // Image URL to process
	MonitoringMeta monitoring.Meta       // Monitoring metadata
}

// PrepareRequest extracts image url and processing options from request URL and verifies them
func NewRequest[P Constructor[O], O Options](
	handler Context, // or, essentially, instance
	constructor P,
	imageRequest *http.Request,
	config *Config,
	reqID string,
	rw http.ResponseWriter,
) (*Request, O, error) {
	// let's extract signature and valid request path from a request
	path, signature, err := splitPathSignature(imageRequest, config)
	if err != nil {
		return nil, *new(O), err
	}

	// verify the signature (if any)
	if err = security.VerifySignature(signature, path); err != nil {
		return nil, *new(O), ierrors.Wrap(err, 0, ierrors.WithCategory(CategorySecurity))
	}

	// parse image url and processing options
	po, imageURL, err := constructor.ParsePath(path, imageRequest.Header)
	if err != nil {
		return nil, *new(O), ierrors.Wrap(err, 0, ierrors.WithCategory(CategoryPathParsing))
	}

	mm := constructor.CreateMeta(imageRequest.Context(), imageURL, po)

	// verify that image URL came from the valid source
	err = security.VerifySourceURL(imageURL)
	if err != nil {
		return nil, *new(O), ierrors.Wrap(err, 0, ierrors.WithCategory(CategorySecurity))
	}

	return &Request{
		Context:        handler,
		Config:         config,
		ID:             reqID,
		Req:            imageRequest,
		ResponseWriter: rw,
		HeaderWriter:   handler.HeaderWriter().NewRequest(),
		ImageURL:       imageURL,
		MonitoringMeta: mm,
	}, po, nil
}

// MakeDownloadOptions creates [imagedata.DownloadOptions]
// from image request headers and security options.
func (r *Request) MakeDownloadOptions(
	h http.Header,
	secops security.Options,
) imagedata.DownloadOptions {
	return imagedata.DownloadOptions{
		Header:         h,
		MaxSrcFileSize: secops.MaxSrcFileSize,
	}
}

// AcquireProcessingSem acquires the processing semaphore.
// It allows as many concurrent processing requests as workers are configured.
func (r *Request) AcquireProcessingSem(ctx context.Context) (context.CancelFunc, error) {
	defer monitoring.StartQueueSegment(ctx)()

	sem := r.Context.Semaphores()

	// Acquire queue semaphore (if enabled)
	releaseQueueSem, err := sem.AcquireQueue()
	if err != nil {
		return nil, err
	}
	// Defer releasing the queue semaphore since we'll exit the queue on return
	defer releaseQueueSem()

	// Acquire processing semaphore
	releaseProcessingSem, err := sem.AcquireProcessing(ctx)
	if err != nil {
		// We don't actually need to check timeout here,
		// but it's an easy way to check if this is an actual timeout
		// or the request was canceled
		if terr := server.CheckTimeout(ctx); terr != nil {
			return nil, ierrors.Wrap(terr, 0, ierrors.WithCategory(CategoryTimeout))
		}

		// We should never reach this line as err could be only ctx.Err()
		// and we've already checked for it. But beter safe than sorry
		return nil, ierrors.Wrap(err, 0, ierrors.WithCategory(CategoryQueue))
	}

	return releaseProcessingSem, nil
}

// MakeImageRequestHeaders creates headers for the image request
func (r *Request) MakeImageRequestHeaders() http.Header {
	h := make(http.Header)

	// If ETag is enabled, we forward If-None-Match header
	if r.Config.ETagEnabled {
		h.Set(httpheaders.IfNoneMatch, r.Req.Header.Get(httpheaders.IfNoneMatch))
	}

	// If LastModified is enabled, we forward If-Modified-Since header
	if r.Config.LastModifiedEnabled {
		h.Set(httpheaders.IfModifiedSince, r.Req.Header.Get(httpheaders.IfModifiedSince))
	}

	return h
}

// FetchImage downloads the source image asynchronously
func (r *Request) FetchImage(
	ctx context.Context,
	do imagedata.DownloadOptions,
) (imagedata.ImageData, http.Header, error) {
	do.DownloadFinished = monitoring.StartDownloadingSegment(ctx, r.MonitoringMeta.Filter(
		monitoring.MetaSourceImageURL,
		monitoring.MetaSourceImageOrigin,
	))

	var err error

	if r.Config.CookiePassthrough {
		do.CookieJar, err = cookies.JarFromRequest(r.Req)
		if err != nil {
			return nil, nil, ierrors.Wrap(err, 0, ierrors.WithCategory(CategoryDownload))
		}
	}

	return r.Context.ImageDataFactory().DownloadAsync(ctx, r.ImageURL, "source image", do)
}

// WrapDownloadingErr wraps original error to download error
func (r *Request) WrapDownloadingErr(originalErr error) *ierrors.Error {
	err := ierrors.Wrap(originalErr, 0, ierrors.WithCategory(CategoryDownload))

	// we report this error only if enabled
	if r.Config.ReportDownloadingErrors {
		err = ierrors.Wrap(err, 0, ierrors.WithShouldReport(true))
	}

	return err
}
