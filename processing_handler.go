package main

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	queueSem      *semaphore.Weighted
	processingSem *semaphore.Weighted
)

func initProcessingHandler() {
	if config.RequestsQueueSize > 0 {
		queueSem = semaphore.NewWeighted(int64(config.RequestsQueueSize + config.Workers))
	}

	processingSem = semaphore.NewWeighted(int64(config.Workers))
}

// writeOriginContentLengthDebugHeader writes the X-Origin-Content-Length header to the response.
func writeOriginContentLengthDebugHeader(rw http.ResponseWriter, originData imagedata.ImageData) error {
	if !config.EnableDebugHeaders {
		return nil
	}

	size, err := originData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryImageDataSize))
	}

	rw.Header().Set(httpheaders.XOriginContentLength, strconv.Itoa(size))

	return nil
}

// writeDebugHeaders writes debug headers (X-Origin-*, X-Result-*) to the response
func writeDebugHeaders(rw http.ResponseWriter, result *processing.Result, originData imagedata.ImageData) error {
	if !config.EnableDebugHeaders || result == nil {
		return nil
	}

	rw.Header().Set(httpheaders.XOriginWidth, strconv.Itoa(result.OriginWidth))
	rw.Header().Set(httpheaders.XOriginHeight, strconv.Itoa(result.OriginHeight))
	rw.Header().Set(httpheaders.XResultWidth, strconv.Itoa(result.ResultWidth))
	rw.Header().Set(httpheaders.XResultHeight, strconv.Itoa(result.ResultHeight))

	return writeOriginContentLengthDebugHeader(rw, originData)
}

func respondWithImage(reqID string, r *http.Request, rw http.ResponseWriter, statusCode int, resultData imagedata.ImageData, po *options.ProcessingOptions, originURL string, originData imagedata.ImageData, hw *headerwriter.Request) error {
	// We read the size of the image data here, so we can set Content-Length header.
	// This indireclty ensures that the image data is fully read from the source, no
	// errors happened.
	resultSize, err := resultData.Size()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryImageDataSize))
	}

	hw.SetContentType(resultData.Format().Mime())
	hw.SetContentLength(resultSize)
	hw.SetContentDisposition(
		originURL,
		po.Filename,
		resultData.Format().Ext(),
		"",
		po.ReturnAttachment,
	)
	hw.SetExpires(po.Expires)
	hw.SetLastModified()
	hw.SetVary()
	hw.SetCanonical()
	hw.SetETag()

	hw.Write(rw)

	rw.WriteHeader(statusCode)

	_, err = io.Copy(rw, resultData.Reader())

	var ierr *ierrors.Error
	if err != nil {
		ierr = newResponseWriteError(err)

		if config.ReportIOErrors {
			return ierrors.Wrap(ierr, 0, ierrors.WithCategory(categoryIO), ierrors.WithShouldReport(true))
		}
	}

	server.LogResponse(
		reqID, r, statusCode, ierr,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)

	return nil
}

func respondWithNotModified(reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string, hw *headerwriter.Request) {
	hw.SetExpires(po.Expires)
	hw.SetVary()
	hw.SetETag()
	hw.Write(rw)

	rw.WriteHeader(http.StatusNotModified)
	server.LogResponse(
		reqID, r, http.StatusNotModified, nil,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)
}

func callHandleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) error {
	// NOTE: This is temporary, will be moved level up at once
	hwc, err := headerwriter.NewDefaultConfig().LoadFromEnv()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	hw, err := headerwriter.New(hwc)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	sc, err := stream.NewDefaultConfig().LoadFromEnv()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	stream, err := stream.New(sc, hw, imagedata.Fetcher)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryConfig))
	}

	return handleProcessing(reqID, rw, r, hw, stream)
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request, hw *headerwriter.Writer, stream *stream.Handler) error {
	stats.IncRequestsInProgress()
	defer stats.DecRequestsInProgress()

	ctx := r.Context()

	path := r.RequestURI
	if queryStart := strings.IndexByte(path, '?'); queryStart >= 0 {
		path = path[:queryStart]
	}

	if len(config.PathPrefix) > 0 {
		path = strings.TrimPrefix(path, config.PathPrefix)
	}

	path = strings.TrimPrefix(path, "/")
	signature := ""

	if signatureEnd := strings.IndexByte(path, '/'); signatureEnd > 0 {
		signature = path[:signatureEnd]
		path = path[signatureEnd:]
	} else {
		return ierrors.Wrap(
			newInvalidURLErrorf(http.StatusNotFound, "Invalid path: %s", path), 0,
			ierrors.WithCategory(categoryPathParsing),
		)
	}

	path = fixPath(path)

	if err := security.VerifySignature(signature, path); err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categorySecurity))
	}

	po, imageURL, err := options.ParsePath(path, r.Header)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryPathParsing))
	}

	var imageOrigin any
	if u, uerr := url.Parse(imageURL); uerr == nil {
		imageOrigin = u.Scheme + "://" + u.Host
	}

	errorreport.SetMetadata(r, "Source Image URL", imageURL)
	errorreport.SetMetadata(r, "Source Image Origin", imageOrigin)
	errorreport.SetMetadata(r, "Processing Options", po)

	monitoringMeta := monitoring.Meta{
		monitoring.MetaSourceImageURL:    imageURL,
		monitoring.MetaSourceImageOrigin: imageOrigin,
		monitoring.MetaProcessingOptions: po.Diff().Flatten(),
	}

	monitoring.SetMetadata(ctx, monitoringMeta)

	err = security.VerifySourceURL(imageURL)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categorySecurity))
	}

	if po.Raw {
		return stream.Execute(ctx, r, imageURL, reqID, po, rw)
	}

	// SVG is a special case. Though saving to svg is not supported, SVG->SVG is.
	if !vips.SupportsSave(po.Format) && po.Format != imagetype.Unknown && po.Format != imagetype.SVG {
		return ierrors.Wrap(newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Resulting image format is not supported: %s", po.Format,
		), 0, ierrors.WithCategory(categoryPathParsing))
	}

	imgRequestHeader := make(http.Header)

	if config.ETagEnabled {
		imgRequestHeader.Set(httpheaders.IfNoneMatch, r.Header.Get(httpheaders.IfNoneMatch))
	}

	if config.LastModifiedEnabled {
		imgRequestHeader.Set(httpheaders.IfModifiedSince, r.Header.Get(httpheaders.IfModifiedSince))
	}

	if queueSem != nil {
		acquired := queueSem.TryAcquire(1)
		if !acquired {
			panic(newTooManyRequestsError())
		}
		defer queueSem.Release(1)
	}

	// The heavy part starts here, so we need to restrict worker number
	err = func() error {
		defer monitoring.StartQueueSegment(ctx)()

		err = processingSem.Acquire(ctx, 1)
		if err != nil {
			// We don't actually need to check timeout here,
			// but it's an easy way to check if this is an actual timeout
			// or the request was canceled
			if terr := server.CheckTimeout(ctx); terr != nil {
				return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
			}

			// We should never reach this line as err could be only ctx.Err()
			// and we've already checked for it. But beter safe than sorry

			return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryQueue))
		}

		return nil
	}()
	if err != nil {
		return err
	}
	defer processingSem.Release(1)

	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	statusCode := http.StatusOK

	originData, originHeaders, err := func() (imagedata.ImageData, http.Header, error) {
		downloadFinished := monitoring.StartDownloadingSegment(ctx, monitoringMeta.Filter(
			monitoring.MetaSourceImageURL,
			monitoring.MetaSourceImageOrigin,
		))

		downloadOpts := imagedata.DownloadOptions{
			Header:           imgRequestHeader,
			CookieJar:        nil,
			MaxSrcFileSize:   po.SecurityOptions.MaxSrcFileSize,
			DownloadFinished: downloadFinished,
		}

		if config.CookiePassthrough {
			downloadOpts.CookieJar, err = cookies.JarFromRequest(r)
			if err != nil {
				return nil, nil, ierrors.Wrap(err, 0, ierrors.WithCategory(categoryDownload))
			}
		}

		return imagedata.DownloadAsync(ctx, imageURL, "source image", downloadOpts)
	}()

	var nmErr imagefetcher.NotModifiedError

	switch {
	case err == nil:
		defer originData.Close()

	case errors.As(err, &nmErr):
		hwr := hw.NewRequest(nmErr.Headers(), imageURL)

		respondWithNotModified(reqID, r, rw, po, imageURL, hwr)
		return nil

	default:
		// This may be a request timeout error or a request cancelled error.
		// Check it before moving further
		if terr := server.CheckTimeout(ctx); terr != nil {
			return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
		}

		ierr := ierrors.Wrap(err, 0, ierrors.WithCategory(categoryDownload))
		if config.ReportDownloadingErrors {
			ierr = ierrors.Wrap(ierr, 0, ierrors.WithShouldReport(true))
		}

		if imagedata.FallbackImage == nil {
			return ierr
		}

		// Just send error
		monitoring.SendError(ctx, categoryDownload, ierr)

		// We didn't return, so we have to report error
		if ierr.ShouldReport() {
			errorreport.Report(ierr, r)
		}

		log.WithField("request_id", reqID).Warningf("Could not load image %s. Using fallback image. %s", imageURL, ierr.Error())

		if config.FallbackImageHTTPCode > 0 {
			statusCode = config.FallbackImageHTTPCode
		} else {
			statusCode = ierr.StatusCode()
		}

		originData = imagedata.FallbackImage
		originHeaders = imagedata.FallbackImageHeaders.Clone()

		if config.FallbackImageTTL > 0 {
			originHeaders.Set("Fallback-Image", "1")
		}
	}

	if terr := server.CheckTimeout(ctx); terr != nil {
		return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
	}

	if !vips.SupportsLoad(originData.Format()) {
		return ierrors.Wrap(newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Source image format is not supported: %s", originData.Format(),
		), 0, ierrors.WithCategory(categoryProcessing))
	}

	result, err := func() (*processing.Result, error) {
		defer monitoring.StartProcessingSegment(ctx, monitoringMeta.Filter(monitoring.MetaProcessingOptions))()
		return processing.ProcessImage(ctx, originData, po)
	}()

	// Let's close resulting image data only if it differs from the source image data
	if result != nil && result.OutData != nil && result.OutData != originData {
		defer result.OutData.Close()
	}

	// First, check if the processing error wasn't caused by an image data error
	if derr := originData.Error(); derr != nil {
		return ierrors.Wrap(derr, 0, ierrors.WithCategory(categoryDownload))
	}

	// If it wasn't, than it was a processing error
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryProcessing))
	}

	if terr := server.CheckTimeout(ctx); terr != nil {
		return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
	}

	hwr := hw.NewRequest(originHeaders, imageURL)

	// Write debug headers. It seems unlogical to move they to headerwriter since they're
	// not used anywhere else.
	err = writeDebugHeaders(rw, result, originData)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryImageDataSize))
	}

	err = respondWithImage(reqID, r, rw, statusCode, result.OutData, po, imageURL, originData, hwr)
	if err != nil {
		return err
	}

	return nil
}
