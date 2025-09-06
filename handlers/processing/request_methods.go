package processing

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/server"
	log "github.com/sirupsen/logrus"
)

// handleDownloadError replaces the image data with fallback image if needed
func handleDownloadError(
	ctx context.Context,
	r *request,
	originalErr error,
) (imagedata.ImageData, int, error) {
	err := r.WrapDownloadingErr(originalErr)

	// If there is no fallback image configured, just return the error
	data, headers := getFallbackImage(ctx, r)
	if data == nil {
		return nil, 0, err
	}

	// Just send error
	monitoring.SendError(ctx, handlers.CategoryDownload, err)

	// We didn't return, so we have to report error
	if err.ShouldReport() {
		errorreport.Report(err, r.Req)
	}

	log.
		WithField("request_id", r.ID).
		Warningf("Could not load image %s. Using fallback image. %s", r.ImageURL, err.Error())

	var statusCode int

	// Set status code if needed
	if r.Config.FallbackImageHTTPCode > 0 {
		statusCode = r.Config.FallbackImageHTTPCode
	} else {
		statusCode = err.StatusCode()
	}

	// Fallback image should have exact FallbackImageTTL lifetime
	headers.Del(httpheaders.Expires)
	headers.Del(httpheaders.LastModified)

	r.HeaderWriter.SetOriginHeaders(headers)
	r.HeaderWriter.SetIsFallbackImage()

	return data, statusCode, nil
}

// getFallbackImage returns fallback image if any
func getFallbackImage(
	ctx context.Context,
	r *request,
) (imagedata.ImageData, http.Header) {
	if r.Handler.fallbackImage == nil {
		return nil, nil
	}

	data, h, err := r.Handler.fallbackImage.Get(ctx, r.Options)
	if err != nil {
		log.Warning(err.Error())

		if ierr := r.WrapDownloadingErr(err); ierr.ShouldReport() {
			errorreport.Report(ierr, r.Req)
		}

		return nil, nil
	}

	return data, h
}

// processImage calls actual image processing
func processImage(
	ctx context.Context,
	r *request,
	originData imagedata.ImageData,
) (*processing.Result, error) {
	defer monitoring.StartProcessingSegment(
		ctx,
		r.MonitoringMeta.Filter(monitoring.MetaProcessingOptions),
	)()
	return processing.ProcessImage(ctx, originData, r.Options, r.Handler.watermarkImage)
}

// writeDebugHeaders writes debug headers (X-Origin-*, X-Result-*) to the response
func writeDebugHeaders(
	r *request,
	result *processing.Result,
	originData imagedata.ImageData,
) error {
	if !r.Config.EnableDebugHeaders {
		return nil
	}

	if result != nil {
		r.ResponseWriter.Header().Set(httpheaders.XOriginWidth, strconv.Itoa(result.OriginWidth))
		r.ResponseWriter.Header().Set(httpheaders.XOriginHeight, strconv.Itoa(result.OriginHeight))
		r.ResponseWriter.Header().Set(httpheaders.XResultWidth, strconv.Itoa(result.ResultWidth))
		r.ResponseWriter.Header().Set(httpheaders.XResultHeight, strconv.Itoa(result.ResultHeight))
	}

	// Try to read origin image size
	size, err := originData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryImageDataSize))
	}

	r.ResponseWriter.Header().Set(httpheaders.XOriginContentLength, strconv.Itoa(size))

	return nil
}

// respondWithNotModified writes not-modified response
func respondWithNotModified(r *request) error {
	r.HeaderWriter.SetExpires(r.Options.Expires)
	r.HeaderWriter.SetVary()

	if r.Config.LastModifiedEnabled {
		r.HeaderWriter.Passthrough(httpheaders.LastModified)
	}

	if r.Config.ETagEnabled {
		r.HeaderWriter.Passthrough(httpheaders.Etag)
	}

	r.HeaderWriter.Write(r.ResponseWriter)

	r.ResponseWriter.WriteHeader(http.StatusNotModified)

	server.LogResponse(
		r.ID, r.Req, http.StatusNotModified, nil,
		log.Fields{
			"image_url":          r.ImageURL,
			"processing_options": r.Options,
		},
	)

	return nil
}

func respondWithImage(r *request, statusCode int, resultData imagedata.ImageData) error {
	// We read the size of the image data here, so we can set Content-Length header.
	// This indireclty ensures that the image data is fully read from the source, no
	// errors happened.
	resultSize, err := resultData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(handlers.CategoryImageDataSize))
	}

	r.HeaderWriter.SetContentType(resultData.Format().Mime())
	r.HeaderWriter.SetContentLength(resultSize)
	r.HeaderWriter.SetContentDisposition(
		r.ImageURL,
		r.Options.Filename,
		resultData.Format().Ext(),
		"",
		r.Options.ReturnAttachment,
	)
	r.HeaderWriter.SetExpires(r.Options.Expires)
	r.HeaderWriter.SetVary()
	r.HeaderWriter.SetCanonical(r.ImageURL)

	if r.Config.LastModifiedEnabled {
		r.HeaderWriter.Passthrough(httpheaders.LastModified)
	}

	if r.Config.ETagEnabled {
		r.HeaderWriter.Passthrough(httpheaders.Etag)
	}

	r.HeaderWriter.Write(r.ResponseWriter)

	r.ResponseWriter.WriteHeader(statusCode)

	_, err = io.Copy(r.ResponseWriter, resultData.Reader())

	var ierr *ierrors.Error
	if err != nil {
		ierr = handlers.NewResponseWriteError(err)

		if r.Config.ReportIOErrors {
			return ierrors.Wrap(
				ierr, 0,
				ierrors.WithCategory(handlers.CategoryIO),
				ierrors.WithShouldReport(true),
			)
		}
	}

	server.LogResponse(
		r.ID, r.Req, statusCode, ierr,
		log.Fields{
			"image_url":          r.ImageURL,
			"processing_options": r.Options,
		},
	)

	return nil
}
