package handlers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/structdiff"
)

// Handler defines the interface for request handler that can be used to create a [Request].
type Handler[O structdiff.Diffable] interface {
	HeaderWriter() *headerwriter.Writer
	Semaphores() *semaphores.Semaphores
	ImageData() *imagedata.Factory

	// ParsePath parses the request path and headers
	// and returns the processing options and image URL
	ParsePath(path string, headers http.Header) (O, string, error)

	// SetMonitoringMeta sets the monitoring metadata for the request.
	// It should return the metadata it set.
	SetMonitoringMeta(ctx context.Context, imageURL, imageOrigin string, po O) monitoring.Meta
}

// Request holds the parameters and state for a single request execution.
// It also exposes common request execution logic as methods.
type Request[H Handler[O], O structdiff.Diffable] struct {
	Handler        H                     // Request handler
	Config         *Config               // Handler configuration
	ID             string                // Request ID
	Req            *http.Request         // Original HTTP request
	ResponseWriter http.ResponseWriter   // HTTP response writer
	HeaderWriter   *headerwriter.Request // Header writer request
	Options        O                     // Processing options
	ImageURL       string                // Image URL to process
	MonitoringMeta monitoring.Meta       // Monitoring metadata
}

// NewRequest parses HTTP request and creates a new [Request] instance.
func NewRequest[H Handler[O], O structdiff.Diffable](
	handler H,
	config *Config,
	reqID string,
	req *http.Request,
	rw http.ResponseWriter,
) (*Request[H, O], error) {
	// let's extract signature and valid request path from a request
	path, signature, err := splitPathSignature(req)
	if err != nil {
		return nil, err
	}

	// verify the signature (if any)
	if err = security.VerifySignature(signature, path); err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithCategory(CategorySecurity))
	}

	// parse image url and processing options
	po, imageURL, err := handler.ParsePath(path, req.Header)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithCategory(CategoryPathParsing))
	}

	// get image origin and set monitoring meta
	imageOrigin := imageOrigin(imageURL)
	mm := handler.SetMonitoringMeta(req.Context(), imageURL, imageOrigin, po)

	// verify that image URL came from the valid source
	err = security.VerifySourceURL(imageURL)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithCategory(CategorySecurity))
	}

	return &Request[H, O]{
		Handler:        handler,
		Config:         config,
		ID:             reqID,
		Req:            req,
		ResponseWriter: rw,
		HeaderWriter:   handler.HeaderWriter().NewRequest(),
		Options:        po,
		ImageURL:       imageURL,
		MonitoringMeta: mm,
	}, nil
}

// imageOrigin extracts image origin from URL
func imageOrigin(imageURL string) string {
	if u, uerr := url.Parse(imageURL); uerr == nil {
		return u.Scheme + "://" + u.Host
	}

	return ""
}

// MakeDownloadOptions creates [imagedata.DownloadOptions]
// from image request headers and security options.
func (r *Request[H, O]) MakeDownloadOptions(
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
func (r *Request[H, O]) AcquireProcessingSem(ctx context.Context) (context.CancelFunc, error) {
	defer monitoring.StartQueueSegment(ctx)()

	sem := r.Handler.Semaphores()

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
func (r *Request[H, O]) MakeImageRequestHeaders() http.Header {
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
func (r *Request[H, O]) FetchImage(
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

	return r.Handler.ImageData().DownloadAsync(ctx, r.ImageURL, "source image", do)
}

// WrapDownloadingErr wraps original error to download error
func (r *Request[H, O]) WrapDownloadingErr(originalErr error) *ierrors.Error {
	err := ierrors.Wrap(originalErr, 0, ierrors.WithCategory(CategoryDownload))

	// we report this error only if enabled
	if r.Config.ReportDownloadingErrors {
		err = ierrors.Wrap(err, 0, ierrors.WithShouldReport(true))
	}

	return err
}
