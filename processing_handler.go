package main

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/svg"
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

	if config.AutoWebp || config.EnforceWebp || config.AutoAvif || config.EnforceAvif {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "Sec-CH-DPR", "DPR", "Sec-CH-Width", "Width")
	}

	headerVaryValue = strings.Join(vary, ", ")
}

func setCacheControl(rw http.ResponseWriter, force *time.Time, originHeaders map[string]string) {
	ttl := -1

	if _, ok := originHeaders["Fallback-Image"]; ok && config.FallbackImageTTL > 0 {
		ttl = config.FallbackImageTTL
	}

	if force != nil && (ttl < 0 || force.Before(time.Now().Add(time.Duration(ttl)*time.Second))) {
		ttl = imath.Min(config.TTL, imath.Max(0, int(time.Until(*force).Seconds())))
	}

	if config.CacheControlPassthrough && ttl < 0 && originHeaders != nil {
		if val, ok := originHeaders["Cache-Control"]; ok && len(val) > 0 {
			rw.Header().Set("Cache-Control", val)
			return
		}

		if val, ok := originHeaders["Expires"]; ok && len(val) > 0 {
			if t, err := time.Parse(http.TimeFormat, val); err == nil {
				ttl = imath.Max(0, int(time.Until(t).Seconds()))
			}
		}
	}

	if ttl < 0 {
		ttl = config.TTL
	}

	if ttl > 0 {
		rw.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", ttl))
	} else {
		rw.Header().Set("Cache-Control", "no-cache")
	}
}

func setLastModified(rw http.ResponseWriter, originHeaders map[string]string) {
	if config.LastModifiedEnabled {
		if val, ok := originHeaders["Last-Modified"]; ok && len(val) != 0 {
			rw.Header().Set("Last-Modified", val)
		}
	}
}

func setVary(rw http.ResponseWriter) {
	if len(headerVaryValue) > 0 {
		rw.Header().Set("Vary", headerVaryValue)
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

func respondWithImage(reqID string, r *http.Request, rw http.ResponseWriter, statusCode int, resultData *imagedata.ImageData, po *options.ProcessingOptions, originURL string, originData *imagedata.ImageData) {
	var contentDisposition string
	if len(po.Filename) > 0 {
		contentDisposition = resultData.Type.ContentDisposition(po.Filename, po.ReturnAttachment)
	} else {
		contentDisposition = resultData.Type.ContentDispositionFromURL(originURL, po.ReturnAttachment)
	}

	rw.Header().Set("Content-Type", resultData.Type.Mime())
	rw.Header().Set("Content-Disposition", contentDisposition)

	setCacheControl(rw, po.Expires, originData.Headers)
	setLastModified(rw, originData.Headers)
	setVary(rw)
	setCanonical(rw, originURL)

	if config.EnableDebugHeaders {
		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(originData.Data)))
		rw.Header().Set("X-Origin-Width", resultData.Headers["X-Origin-Width"])
		rw.Header().Set("X-Origin-Height", resultData.Headers["X-Origin-Height"])
		rw.Header().Set("X-Result-Width", resultData.Headers["X-Result-Width"])
		rw.Header().Set("X-Result-Height", resultData.Headers["X-Result-Height"])
	}

	rw.Header().Set("Content-Security-Policy", "script-src 'none'")

	rw.Header().Set("Content-Length", strconv.Itoa(len(resultData.Data)))
	rw.WriteHeader(statusCode)
	_, err := rw.Write(resultData.Data)

	var ierr *ierrors.Error
	if err != nil {
		ierr = ierrors.New(statusCode, fmt.Sprintf("Failed to write response: %s", err), "Failed to write response")
		ierr.Unexpected = true

		if config.ReportIOErrors {
			sendErr(r.Context(), "IO", ierr)
			errorreport.Report(ierr, r)
		}
	}

	router.LogResponse(
		reqID, r, statusCode, ierr,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)
}

func respondWithNotModified(reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string, originHeaders map[string]string) {
	setCacheControl(rw, po.Expires, originHeaders)
	setVary(rw)

	rw.WriteHeader(304)
	router.LogResponse(
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
		switch ierr.StatusCode {
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

func sendErrAndPanic(ctx context.Context, errType string, err error) {
	sendErr(ctx, errType, err)
	panic(err)
}

func checkErr(ctx context.Context, errType string, err error) {
	if err == nil {
		return
	}
	sendErrAndPanic(ctx, errType, err)
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) {
	stats.IncRequestsInProgress()
	defer stats.DecRequestsInProgress()

	ctx := r.Context()

	path := r.RequestURI

	if !config.SourceURLQueryPassthrough {
		if queryStart := strings.IndexByte(path, '?'); queryStart >= 0 {
			path = path[:queryStart]
		}
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
		sendErrAndPanic(ctx, "path_parsing", ierrors.New(
			404, fmt.Sprintf("Invalid path: %s", path), "Invalid URL",
		))
	}

	path = fixPath(path)

	if err := security.VerifySignature(signature, path); err != nil {
		sendErrAndPanic(ctx, "security", ierrors.New(403, err.Error(), "Forbidden"))
	}

	po, imageURL, err := options.ParsePath(path, r.Header)
	checkErr(ctx, "path_parsing", err)

	errorreport.SetMetadata(r, "Source Image URL", imageURL)
	errorreport.SetMetadata(r, "Processing Options", po)

	metrics.SetMetadata(ctx, "imgproxy.source_image_url", imageURL)
	metrics.SetMetadata(ctx, "imgproxy.processing_options", po)

	err = security.VerifySourceURL(imageURL)
	checkErr(ctx, "security", err)

	if po.Raw {
		streamOriginImage(ctx, reqID, r, rw, po, imageURL)
		return
	}

	// SVG is a special case. Though saving to svg is not supported, SVG->SVG is.
	if !vips.SupportsSave(po.Format) && po.Format != imagetype.Unknown && po.Format != imagetype.SVG {
		sendErrAndPanic(ctx, "path_parsing", ierrors.New(
			422,
			fmt.Sprintf("Resulting image format is not supported: %s", po.Format),
			"Invalid URL",
		))
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
			panic(ierrors.New(429, "Too many requests", "Too many requests"))
		}
		defer queueSem.Release(1)
	}

	// The heavy part starts here, so we need to restrict worker number
	func() {
		defer metrics.StartQueueSegment(ctx)()

		err = processingSem.Acquire(ctx, 1)
		if err != nil {
			// We don't actually need to check timeout here,
			// but it's an easy way to check if this is an actual timeout
			// or the request was canceled
			checkErr(ctx, "queue", router.CheckTimeout(ctx))
			// We should never reach this line as err could be only ctx.Err()
			// and we've already checked for it. But beter safe than sorry
			sendErrAndPanic(ctx, "queue", err)
		}
	}()
	defer processingSem.Release(1)

	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	statusCode := http.StatusOK

	originData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartDownloadingSegment(ctx)()

		downloadOpts := imagedata.DownloadOptions{
			Header:    imgRequestHeader,
			CookieJar: nil,
		}

		if config.CookiePassthrough {
			downloadOpts.CookieJar, err = cookies.JarFromRequest(r)
			checkErr(ctx, "download", err)
		}

		return imagedata.Download(ctx, imageURL, "source image", downloadOpts, po.SecurityOptions)
	}()

	if err == nil {
		defer originData.Close()
	} else if nmErr, ok := err.(*imagedata.ErrorNotModified); ok {
		if config.ETagEnabled && len(etagHandler.ImageEtagExpected()) != 0 {
			rw.Header().Set("ETag", etagHandler.GenerateExpectedETag())
		}
		respondWithNotModified(reqID, r, rw, po, imageURL, nmErr.Headers)
		return
	} else {
		// This may be a request timeout error or a request cancelled error.
		// Check it before moving further
		checkErr(ctx, "timeout", router.CheckTimeout(ctx))

		ierr := ierrors.Wrap(err, 0)
		ierr.Unexpected = ierr.Unexpected || config.ReportDownloadingErrors

		sendErr(ctx, "download", ierr)

		if imagedata.FallbackImage == nil {
			panic(ierr)
		}

		// We didn't panic, so the error is not reported.
		// Report it now
		if ierr.Unexpected {
			errorreport.Report(ierr, r)
		}

		log.WithField("request_id", reqID).Warningf("Could not load image %s. Using fallback image. %s", imageURL, ierr.Error())

		if config.FallbackImageHTTPCode > 0 {
			statusCode = config.FallbackImageHTTPCode
		} else {
			statusCode = ierr.StatusCode
		}

		originData = imagedata.FallbackImage
	}

	checkErr(ctx, "timeout", router.CheckTimeout(ctx))

	if config.ETagEnabled && statusCode == http.StatusOK {
		imgDataMatch := etagHandler.SetActualImageData(originData)

		rw.Header().Set("ETag", etagHandler.GenerateActualETag())

		if imgDataMatch && etagHandler.ProcessingOptionsMatch() {
			respondWithNotModified(reqID, r, rw, po, imageURL, originData.Headers)
			return
		}
	}

	checkErr(ctx, "timeout", router.CheckTimeout(ctx))

	// Skip processing svg with unknown or the same destination imageType
	// if it's not forced by AlwaysRasterizeSvg option
	// Also skip processing if the format is in SkipProcessingFormats
	shouldSkipProcessing := (originData.Type == po.Format || po.Format == imagetype.Unknown) &&
		(slices.Contains(po.SkipProcessingFormats, originData.Type) ||
			originData.Type == imagetype.SVG && !config.AlwaysRasterizeSvg)

	if shouldSkipProcessing {
		if originData.Type == imagetype.SVG && config.SanitizeSvg {
			sanitized, svgErr := svg.Sanitize(originData)
			checkErr(ctx, "svg_processing", svgErr)

			// Since we'll replace origin data, it's better to close it to return
			// it's buffer to the pool
			originData.Close()

			originData = sanitized

		}

		respondWithImage(reqID, r, rw, statusCode, originData, po, imageURL, originData)
		return
	}

	if !vips.SupportsLoad(originData.Type) {
		sendErrAndPanic(ctx, "processing", ierrors.New(
			422,
			fmt.Sprintf("Source image format is not supported: %s", originData.Type),
			"Invalid URL",
		))
	}

	// At this point we can't allow requested format to be SVG as we can't save SVGs
	if po.Format == imagetype.SVG {
		sendErrAndPanic(ctx, "processing", ierrors.New(
			422, "Resulting image format is not supported: svg", "Invalid URL",
		))
	}

	// We're going to rasterize SVG. Since librsvg lacks the support of some SVG
	// features, we're going to replace them to minimize rendering error
	if originData.Type == imagetype.SVG && config.SvgFixUnsupported {
		fixed, changed, svgErr := svg.FixUnsupported(originData)
		checkErr(ctx, "svg_processing", svgErr)

		if changed {
			// Since we'll replace origin data, it's better to close it to return
			// it's buffer to the pool
			originData.Close()

			originData = fixed
		}
	}

	resultData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartProcessingSegment(ctx)()
		return processing.ProcessImage(ctx, originData, po)
	}()
	checkErr(ctx, "processing", err)

	defer resultData.Close()

	checkErr(ctx, "timeout", router.CheckTimeout(ctx))

	respondWithImage(reqID, r, rw, statusCode, resultData, po, imageURL, originData)
}
