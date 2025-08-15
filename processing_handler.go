package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	queueSem      *semaphore.Weighted
	processingSem *semaphore.Weighted

	headerVaryValue string
)

func initProcessingHandler() {
	if config.RequestsQueueSize > 0 {
		queueSem = semaphore.NewWeighted(int64(config.RequestsQueueSize + config.Workers))
	}

	processingSem = semaphore.NewWeighted(int64(config.Workers))

	vary := make([]string, 0)

	if config.AutoWebp ||
		config.EnforceWebp ||
		config.AutoAvif ||
		config.EnforceAvif ||
		config.AutoJxl ||
		config.EnforceJxl {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	headerVaryValue = strings.Join(vary, ", ")
}

func setCacheControl(rw http.ResponseWriter, force *time.Time, originHeaders http.Header) {
	ttl := -1

	if _, ok := originHeaders["Fallback-Image"]; ok && config.FallbackImageTTL > 0 {
		ttl = config.FallbackImageTTL
	}

	if force != nil && (ttl < 0 || force.Before(time.Now().Add(time.Duration(ttl)*time.Second))) {
		ttl = min(config.TTL, max(0, int(time.Until(*force).Seconds())))
	}

	if config.CacheControlPassthrough && ttl < 0 && originHeaders != nil {
		if val := originHeaders.Get(httpheaders.CacheControl); len(val) > 0 {
			rw.Header().Set(httpheaders.CacheControl, val)
			return
		}

		if val := originHeaders.Get(httpheaders.Expires); len(val) > 0 {
			if t, err := time.Parse(http.TimeFormat, val); err == nil {
				ttl = max(0, int(time.Until(t).Seconds()))
			}
		}
	}

	if ttl < 0 {
		ttl = config.TTL
	}

	if ttl > 0 {
		rw.Header().Set(httpheaders.CacheControl, fmt.Sprintf("max-age=%d, public", ttl))
	} else {
		rw.Header().Set(httpheaders.CacheControl, "no-cache")
	}
}

func setLastModified(rw http.ResponseWriter, originHeaders http.Header) {
	if config.LastModifiedEnabled {
		if val := originHeaders.Get(httpheaders.LastModified); len(val) != 0 {
			rw.Header().Set(httpheaders.LastModified, val)
		}
	}
}

func setVary(rw http.ResponseWriter) {
	if len(headerVaryValue) > 0 {
		rw.Header().Set(httpheaders.Vary, headerVaryValue)
	}
}

func setCanonical(rw http.ResponseWriter, originURL string) {
	if config.SetCanonicalHeader {
		if strings.HasPrefix(originURL, "https://") || strings.HasPrefix(originURL, "http://") {
			linkHeader := fmt.Sprintf(`<%s>; rel="canonical"`, originURL)
			rw.Header().Set("Link", linkHeader)
		}
	}
}

func writeOriginContentLengthDebugHeader(rw http.ResponseWriter, originData imagedata.ImageData) error {
	if !config.EnableDebugHeaders {
		return nil
	}

	size, err := originData.Size()
	if err != nil {
		return ierrors.Wrap(
			err, 0,
			ierrors.WithCategory(categoryImageDataSize),
			ierrors.WithShouldReport(true),
		)
	}

	rw.Header().Set(httpheaders.XOriginContentLength, strconv.Itoa(size))

	return nil
}

func writeDebugHeaders(rw http.ResponseWriter, result *processing.Result) {
	if !config.EnableDebugHeaders || result == nil {
		return
	}

	rw.Header().Set(httpheaders.XOriginWidth, strconv.Itoa(result.OriginWidth))
	rw.Header().Set(httpheaders.XOriginHeight, strconv.Itoa(result.OriginHeight))
	rw.Header().Set(httpheaders.XResultWidth, strconv.Itoa(result.ResultWidth))
	rw.Header().Set(httpheaders.XResultHeight, strconv.Itoa(result.ResultHeight))
}

func respondWithImage(reqID string, r *http.Request, rw http.ResponseWriter, statusCode int, resultData imagedata.ImageData, po *options.ProcessingOptions, originURL string, originHeaders http.Header) error {
	// We read the size of the image data here, so we can set Content-Length header.
	// This indireclty ensures that the image data is fully read from the source, no
	// errors happened.
	resultSize, err := resultData.Size()
	if err != nil {
		return ierrors.Wrap(
			err, 0,
			ierrors.WithCategory(categoryImageDataSize),
			ierrors.WithShouldReport(true),
		)
	}

	contentDisposition := httpheaders.ContentDispositionValue(
		originURL,
		po.Filename,
		resultData.Format().Ext(),
		"",
		po.ReturnAttachment,
	)

	rw.Header().Set(httpheaders.ContentType, resultData.Format().Mime())
	rw.Header().Set(httpheaders.ContentDisposition, contentDisposition)

	setCacheControl(rw, po.Expires, originHeaders)
	setLastModified(rw, originHeaders)
	setVary(rw)
	setCanonical(rw, originURL)

	rw.Header().Set(httpheaders.ContentSecurityPolicy, "script-src 'none'")

	rw.Header().Set(httpheaders.ContentLength, strconv.Itoa(resultSize))
	rw.WriteHeader(statusCode)

	_, err = io.Copy(rw, resultData.Reader())

	var ierr *ierrors.Error
	if err != nil {
		ierr = newResponseWriteError(err)

		if config.ReportIOErrors {
			sendErr(r.Context(), categoryIO, ierr)
			errorreport.Report(ierr, r)
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

func respondWithNotModified(reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string, originHeaders http.Header) {
	setCacheControl(rw, po.Expires, originHeaders)
	setVary(rw)

	rw.WriteHeader(304)
	server.LogResponse(
		reqID, r, 304, nil,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)
}

func sendErr(ctx context.Context, errType string, err error) {
	send := true

	if ierr, ok := err.(*ierrors.Error); ok {
		switch ierr.StatusCode() {
		case http.StatusServiceUnavailable:
			errType = "timeout"
		case 499:
			// Don't need to send a "request cancelled" error
			send = false
		}
	}

	if send {
		metrics.SendError(ctx, errType, err)
	}
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) error {
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

	metricsMeta := metrics.Meta{
		metrics.MetaSourceImageURL:    imageURL,
		metrics.MetaSourceImageOrigin: imageOrigin,
		metrics.MetaProcessingOptions: po.Diff().Flatten(),
	}

	metrics.SetMetadata(ctx, metricsMeta)

	err = security.VerifySourceURL(imageURL)
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categorySecurity))
	}

	if po.Raw {
		streamOriginImage(ctx, reqID, r, rw, po, imageURL)
		return nil
	}

	// SVG is a special case. Though saving to svg is not supported, SVG->SVG is.
	if !vips.SupportsSave(po.Format) && po.Format != imagetype.Unknown && po.Format != imagetype.SVG {
		return ierrors.Wrap(newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Resulting image format is not supported: %s", po.Format,
		), 0, ierrors.WithCategory(categoryPathParsing))
	}

	imgRequestHeader := make(http.Header)

	var etagHandler etag.Handler

	if config.ETagEnabled {
		etagHandler.ParseExpectedETag(r.Header.Get("If-None-Match"))

		if etagHandler.SetActualProcessingOptions(po) {
			if imgEtag := etagHandler.ImageEtagExpected(); len(imgEtag) != 0 {
				imgRequestHeader.Set("If-None-Match", imgEtag)
			}
		}
	}

	if config.LastModifiedEnabled {
		if modifiedSince := r.Header.Get("If-Modified-Since"); len(modifiedSince) != 0 {
			imgRequestHeader.Set("If-Modified-Since", modifiedSince)
		}
	}

	if queueSem != nil {
		acquired := queueSem.TryAcquire(1)
		if !acquired {
			panic(newTooManyRequestsError())
		}
		defer queueSem.Release(1)
	}

	// The heavy part starts here, so we need to restrict worker number

	err = processingSem.Acquire(ctx, 1)
	if err != nil {
		metrics.StartQueueSegment(ctx)()

		// We don't actually need to check timeout here,
		// but it's an easy way to check if this is an actual timeout
		// or the request was canceled
		if terr := server.CheckTimeout(ctx); terr != nil {
			return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
		}

		if err != nil {
			return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryQueue))
		}
	}
	defer processingSem.Release(1)
	metrics.StartQueueSegment(ctx)()

	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	statusCode := http.StatusOK

	originData, originHeaders, err := func() (imagedata.ImageData, http.Header, error) {
		downloadFinished := metrics.StartDownloadingSegment(ctx, metrics.Meta{
			metrics.MetaSourceImageURL:    metricsMeta[metrics.MetaSourceImageURL],
			metrics.MetaSourceImageOrigin: metricsMeta[metrics.MetaSourceImageOrigin],
		})

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
		if config.ETagEnabled && len(etagHandler.ImageEtagExpected()) != 0 {
			rw.Header().Set(httpheaders.Etag, etagHandler.GenerateExpectedETag())
		}

		respondWithNotModified(reqID, r, rw, po, imageURL, nmErr.Headers())
		return nil

	default:
		// This may be a request timeout error or a request cancelled error.
		// Check it before moving further
		if terr := server.CheckTimeout(ctx); terr != nil {
			return ierrors.Wrap(terr, 0, ierrors.WithCategory(categoryTimeout))
		}

		ierr := ierrors.Wrap(err, 0)
		if config.ReportDownloadingErrors {
			ierr = ierrors.Wrap(ierr, 0, ierrors.WithShouldReport(true))
		}

		if ierr != nil {
			metrics.SendError(ctx, categoryDownload, err)
		}

		if imagedata.FallbackImage == nil {
			return ierr
		}

		// Fallback image was present, however, we did not report it
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

	if config.ETagEnabled && statusCode == http.StatusOK {
		imgDataMatch, terr := etagHandler.SetActualImageData(originData, originHeaders)
		if terr == nil {
			rw.Header().Set("ETag", etagHandler.GenerateActualETag())

			if imgDataMatch && etagHandler.ProcessingOptionsMatch() {
				respondWithNotModified(reqID, r, rw, po, imageURL, originHeaders)
				return nil
			}
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
		defer metrics.StartProcessingSegment(ctx, metrics.Meta{
			metrics.MetaProcessingOptions: metricsMeta[metrics.MetaProcessingOptions],
		})()
		return processing.ProcessImage(ctx, originData, po)
	}()

	// Let's close resulting image data only if it differs from the source image data
	if result != nil && result.OutData != nil && result.OutData != originData {
		defer result.OutData.Close()
	}

	if err != nil {
		// First, check if the processing error wasn't caused by an image data error
		if originData.Error() != nil {
			return ierrors.Wrap(originData.Error(), 0, ierrors.WithCategory(categoryDownload))
		}

		// If it wasn't, than it was a processing error
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryProcessing))
	}

	if err := server.CheckTimeout(ctx); err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryTimeout))
	}

	writeDebugHeaders(rw, result)
	if err := writeOriginContentLengthDebugHeader(rw, originData); err != nil {
		return err
	}

	respondWithImage(reqID, r, rw, statusCode, result.OutData, po, imageURL, originHeaders)

	return nil
}
