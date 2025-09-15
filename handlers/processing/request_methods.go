package processing

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/server"
	log "github.com/sirupsen/logrus"
)

// makeImageRequestHeaders creates headers for the image request
func (r *request) makeImageRequestHeaders() http.Header {
	h := make(http.Header)

	// If ETag is enabled, we forward If-None-Match header
	if r.config.ETagEnabled {
		h.Set(httpheaders.IfNoneMatch, r.req.Header.Get(httpheaders.IfNoneMatch))
	}

	// If LastModified is enabled, we forward If-Modified-Since header
	if r.config.LastModifiedEnabled {
		h.Set(httpheaders.IfModifiedSince, r.req.Header.Get(httpheaders.IfModifiedSince))
	}

	return h
}

// acquireWorker acquires the processing worker
func (r *request) acquireWorker(ctx context.Context) (context.CancelFunc, error) {
	defer monitoring.StartQueueSegment(ctx)()

	fn, err := r.Workers().Acquire(ctx)
	if err != nil {
		// We don't actually need to check timeout here,
		// but it's an easy way to check if this is an actual timeout
		// or the request was canceled
		if terr := server.CheckTimeout(ctx); terr != nil {
			return nil, ierrors.Wrap(terr, 0, ierrors.WithCategory(handlers.CategoryTimeout))
		}

		// We should never reach this line as err could be only ctx.Err()
		// and we've already checked for it. But beter safe than sorry
		return nil, ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryQueue))
	}

	return fn, nil
}

// makeDownloadOptions creates a new default download options
func (r *request) makeDownloadOptions(ctx context.Context, h http.Header) imagedata.DownloadOptions {
	downloadFinished := monitoring.StartDownloadingSegment(ctx, r.monitoringMeta.Filter(
		monitoring.MetaSourceImageURL,
		monitoring.MetaSourceImageOrigin,
	))

	return imagedata.DownloadOptions{
		Header:           h,
		MaxSrcFileSize:   r.po.SecurityOptions.MaxSrcFileSize,
		DownloadFinished: downloadFinished,
	}
}

// fetchImage downloads the source image asynchronously
func (r *request) fetchImage(ctx context.Context, do imagedata.DownloadOptions) (imagedata.ImageData, http.Header, error) {
	var err error

	if r.config.CookiePassthrough {
		do.CookieJar, err = cookies.JarFromRequest(r.req)
		if err != nil {
			return nil, nil, ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryDownload))
		}
	}

	return r.ImageDataFactory().DownloadAsync(ctx, r.imageURL, "source image", do)
}

// handleDownloadError replaces the image data with fallback image if needed
func (r *request) handleDownloadError(
	ctx context.Context,
	originalErr error,
) (imagedata.ImageData, int, error) {
	err := r.wrapDownloadingErr(originalErr)

	// If there is no fallback image configured, just return the error
	data, headers := r.getFallbackImage(ctx, r.po)
	if data == nil {
		return nil, 0, err
	}

	// Just send error
	monitoring.SendError(ctx, handlers.CategoryDownload, err)

	// We didn't return, so we have to report error
	if err.ShouldReport() {
		errorreport.Report(err, r.req)
	}

	log.
		WithField("request_id", r.reqID).
		Warningf("Could not load image %s. Using fallback image. %s", r.imageURL, err.Error())

	var statusCode int

	// Set status code if needed
	if r.config.FallbackImageHTTPCode > 0 {
		statusCode = r.config.FallbackImageHTTPCode
	} else {
		statusCode = err.StatusCode()
	}

	// Fallback image should have exact FallbackImageTTL lifetime
	headers.Del(httpheaders.Expires)
	headers.Del(httpheaders.LastModified)

	r.rw.SetOriginHeaders(headers)
	r.rw.SetIsFallbackImage()

	return data, statusCode, nil
}

// getFallbackImage returns fallback image if any
func (r *request) getFallbackImage(
	ctx context.Context,
	po *options.ProcessingOptions,
) (imagedata.ImageData, http.Header) {
	fbi := r.FallbackImage()

	if fbi == nil {
		return nil, nil
	}

	data, h, err := fbi.Get(ctx, po)
	if err != nil {
		log.Warning(err.Error())

		if ierr := r.wrapDownloadingErr(err); ierr.ShouldReport() {
			errorreport.Report(ierr, r.req)
		}

		return nil, nil
	}

	return data, h
}

// processImage calls actual image processing
func (r *request) processImage(ctx context.Context, originData imagedata.ImageData) (*processing.Result, error) {
	defer monitoring.StartProcessingSegment(ctx, r.monitoringMeta.Filter(monitoring.MetaProcessingOptions))()
	return processing.ProcessImage(ctx, originData, r.po, r.WatermarkImage(), r.ProcessingOptionsFactory())
}

// writeDebugHeaders writes debug headers (X-Origin-*, X-Result-*) to the response
func (r *request) writeDebugHeaders(result *processing.Result, originData imagedata.ImageData) error {
	if !r.config.EnableDebugHeaders {
		return nil
	}

	if result != nil {
		r.rw.Header().Set(httpheaders.XOriginWidth, strconv.Itoa(result.OriginWidth))
		r.rw.Header().Set(httpheaders.XOriginHeight, strconv.Itoa(result.OriginHeight))
		r.rw.Header().Set(httpheaders.XResultWidth, strconv.Itoa(result.ResultWidth))
		r.rw.Header().Set(httpheaders.XResultHeight, strconv.Itoa(result.ResultHeight))
	}

	// Try to read origin image size
	size, err := originData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryImageDataSize))
	}

	r.rw.Header().Set(httpheaders.XOriginContentLength, strconv.Itoa(size))

	return nil
}

// respondWithNotModified writes not-modified response
func (r *request) respondWithNotModified() error {
	r.rw.SetExpires(r.po.Expires)
	r.rw.SetVary()

	if r.config.LastModifiedEnabled {
		r.rw.Passthrough(httpheaders.LastModified)
	}

	if r.config.ETagEnabled {
		r.rw.Passthrough(httpheaders.Etag)
	}

	r.rw.WriteHeader(http.StatusNotModified)

	server.LogResponse(
		r.reqID, r.req, http.StatusNotModified, nil,
		log.Fields{
			"image_url":          r.imageURL,
			"processing_options": r.po,
		},
	)

	return nil
}

func (r *request) respondWithImage(statusCode int, resultData imagedata.ImageData) error {
	// We read the size of the image data here, so we can set Content-Length header.
	// This indireclty ensures that the image data is fully read from the source, no
	// errors happened.
	resultSize, err := resultData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryImageDataSize))
	}

	r.rw.SetContentType(resultData.Format().Mime())
	r.rw.SetContentLength(resultSize)
	r.rw.SetContentDisposition(
		r.imageURL,
		r.po.Filename,
		resultData.Format().Ext(),
		"",
		r.po.ReturnAttachment,
	)
	r.rw.SetExpires(r.po.Expires)
	r.rw.SetVary()
	r.rw.SetCanonical(r.imageURL)

	if r.config.LastModifiedEnabled {
		r.rw.Passthrough(httpheaders.LastModified)
	}

	if r.config.ETagEnabled {
		r.rw.Passthrough(httpheaders.Etag)
	}

	r.rw.WriteHeader(statusCode)

	_, err = io.Copy(r.rw, resultData.Reader())

	var ierr *ierrors.Error
	if err != nil {
		ierr = handlers.NewResponseWriteError(err)

		if r.config.ReportIOErrors {
			return ierrors.Wrap(ierr, 0, ierrors.WithCategory(handlers.CategoryIO), ierrors.WithShouldReport(true))
		}
	}

	server.LogResponse(
		r.reqID, r.req, statusCode, ierr,
		log.Fields{
			"image_url":          r.imageURL,
			"processing_options": r.po,
		},
	)

	return nil
}

// wrapDownloadingErr wraps original error to download error
func (r *request) wrapDownloadingErr(originalErr error) *ierrors.Error {
	err := ierrors.Wrap(originalErr, 0, ierrors.WithCategory(handlers.CategoryDownload))

	// we report this error only if enabled
	if r.config.ReportDownloadingErrors {
		err = ierrors.Wrap(err, 0, ierrors.WithShouldReport(true))
	}

	return err
}
