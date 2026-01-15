package processing

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/server"
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
func (r *request) acquireWorker() (context.CancelFunc, errctx.Error) {
	ctx, cancelSpan := r.Monitoring().StartSpan(r.req.Context(), "Queue", nil)
	defer cancelSpan()

	fn, err := r.Workers().Acquire(ctx)
	if err != nil {
		// We don't actually need to check timeout here,
		// but it's an easy way to check if this is an actual timeout
		// or the request was canceled
		if terr := server.CheckTimeout(ctx); terr != nil {
			return nil, terr
		}

		// We should never reach this line as err could be only ctx.Err()
		// and we've already checked for it. But beter safe than sorry
		return nil, errctx.Wrap(err)
	}

	return fn, nil
}

// makeDownloadOptions creates a new default download options
func (r *request) makeDownloadOptions(
	h http.Header,
) (imagedata.DownloadOptions, errctx.Error) {
	jar, err := r.Cookies().JarFromRequest(r.req)
	if err != nil {
		return imagedata.DownloadOptions{}, r.wrapDownloadingErr(err)
	}

	return imagedata.DownloadOptions{
		Header:         h,
		MaxSrcFileSize: r.Security().MaxSrcFileSize(r.opts),
		CookieJar:      jar,
	}, nil
}

// fetchImage downloads the source image asynchronously
func (r *request) fetchImage(
	do imagedata.DownloadOptions,
) (imagedata.ImageData, http.Header, errctx.Error) {
	data, h, err := r.ImageDataFactory().DownloadAsync(
		r.req.Context(),
		r.imageURL,
		"source image",
		do,
	)
	return data, h, r.wrapDownloadingErr(err)
}

// handleDownloadError replaces the image data with fallback image if needed
func (r *request) handleDownloadError(
	err errctx.Error,
) (imagedata.ImageData, int, errctx.Error) {
	// If there is no fallback image configured, just return the error
	data, headers := r.getFallbackImage()
	if data == nil {
		return nil, 0, err
	}

	// Just send error
	r.Monitoring().SendError(r.req.Context(), handlers.ErrCategoryDownload, err)

	// We didn't return, so we have to report error
	if err.ShouldReport() {
		r.ErrorReporter().Report(err, r.req)
	}

	slog.Warn(
		"Could not load image. Using fallback image",
		"request_id", r.reqID,
		"image_url", r.imageURL,
		"error", err.Error(),
	)

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
func (r *request) getFallbackImage() (imagedata.ImageData, http.Header) {
	fbi := r.FallbackImage()

	if fbi == nil {
		return nil, nil
	}

	data, h, err := fbi.Get(r.req.Context(), r.opts)
	if err != nil {
		slog.Warn(err.Error())

		if ierr := r.wrapDownloadingErr(err); ierr.ShouldReport() {
			r.ErrorReporter().Report(ierr, r.req)
		}

		return nil, nil
	}

	return data, h
}

// processImage calls actual image processing
func (r *request) processImage(
	originData imagedata.ImageData,
) (*processing.Result, errctx.Error) {
	ctx, cancelSpan := r.Monitoring().StartSpan(
		r.req.Context(),
		"Processing image",
		r.monitoringMeta.Filter(
			monitoring.MetaOptions,
		),
	)
	defer cancelSpan()

	res, err := r.Processor().ProcessImage(ctx, originData, r.opts)
	return res, errctx.Wrap(err)
}

// writeDebugHeaders writes debug headers (X-Origin-*, X-Result-*) to the response
func (r *request) writeDebugHeaders(
	result *processing.Result,
	originData imagedata.ImageData,
) *server.Error {
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
		return server.NewError(errctx.Wrap(err), handlers.ErrCategoryImageDataSize)
	}

	r.rw.Header().Set(httpheaders.XOriginContentLength, strconv.Itoa(size))

	return nil
}

// respondWithNotModified writes not-modified response
func (r *request) respondWithNotModified() {
	r.rw.SetExpires(r.opts.GetTime(keys.Expires))

	if r.config.LastModifiedEnabled {
		r.rw.Passthrough(httpheaders.LastModified)
	}

	if r.config.ETagEnabled {
		r.rw.Passthrough(httpheaders.Etag)
	}

	r.ClientFeaturesDetector().SetVary(r.rw.Header())

	r.rw.WriteHeader(http.StatusNotModified)

	server.LogResponse(
		r.reqID, r.req, http.StatusNotModified, nil,
		slog.String("image_url", r.imageURL),
		slog.Any("processing_options", r.opts),
	)
}

func (r *request) respondWithImage(statusCode int, resultData imagedata.ImageData) *server.Error {
	// We read the size of the image data here, so we can set Content-Length header.
	// This indirectly ensures that the image data is fully read from the source, no
	// errors happened.
	resultSize, err := resultData.Size()
	if err != nil {
		return server.NewError(errctx.Wrap(err), handlers.ErrCategoryImageDataSize)
	}

	r.rw.SetContentType(resultData.Format().Mime())
	r.rw.SetContentLength(resultSize)
	r.rw.SetContentDisposition(
		r.imageURL,
		r.opts.GetString(keys.Filename, ""),
		resultData.Format().Ext(),
		"",
		r.opts.GetBool(keys.ReturnAttachment, false),
	)
	r.rw.SetExpires(r.opts.GetTime(keys.Expires))
	r.rw.SetCanonical(r.imageURL)

	r.ClientFeaturesDetector().SetVary(r.rw.Header())

	if r.config.LastModifiedEnabled {
		r.rw.Passthrough(httpheaders.LastModified)
	}

	if r.config.ETagEnabled {
		r.rw.Passthrough(httpheaders.Etag)
	}

	r.rw.WriteHeader(statusCode)

	_, err = io.Copy(r.rw, resultData.Reader())

	var ierr errctx.Error
	if err != nil {
		ierr = handlers.NewResponseWriteError(err)

		if r.config.ReportIOErrors {
			return server.NewError(ierr, handlers.ErrCategoryIO)
		}
	}

	server.LogResponse(
		r.reqID, r.req, statusCode, ierr,
		slog.String("image_url", r.imageURL),
		slog.Any("processing_options", r.opts),
	)

	return nil
}

// wrapDownloadingErr wraps original error to download error
func (r *request) wrapDownloadingErr(originalErr error) errctx.Error {
	if originalErr == nil {
		return nil
	}

	var opts []errctx.Option
	// we report this error only if enabled
	if r.config.ReportDownloadingErrors {
		opts = []errctx.Option{errctx.WithShouldReport(true)}
	}

	return errctx.WrapWithStackSkip(originalErr, 1, opts...)
}
