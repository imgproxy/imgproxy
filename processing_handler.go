package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/semaphore"
	"github.com/imgproxy/imgproxy/v3/svg"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	queueSem      *semaphore.Semaphore
	processingSem *semaphore.Semaphore

	headerVaryValue string
)

func initProcessingHandler() {
	if config.RequestsQueueSize > 0 {
		queueSem = semaphore.New(config.RequestsQueueSize + config.Concurrency)
	}

	processingSem = semaphore.New(config.Concurrency)

	vary := make([]string, 0)

	if config.EnableWebpDetection || config.EnforceWebp {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "DPR", "Viewport-Width", "Width")
	}

	headerVaryValue = strings.Join(vary, ", ")
}

func setCacheControl(rw http.ResponseWriter, originHeaders map[string]string) {
	var cacheControl, expires string
	var ttl int

	if config.CacheControlPassthrough && originHeaders != nil {
		if val, ok := originHeaders["Cache-Control"]; ok && len(val) > 0 {
			cacheControl = val
		}
		if val, ok := originHeaders["Expires"]; ok && len(val) > 0 {
			expires = val
		}
	}

	if len(cacheControl) == 0 && len(expires) == 0 {
		ttl = config.TTL
		if _, ok := originHeaders["Fallback-Image"]; ok && config.FallbackImageTTL > 0 {
			ttl = config.FallbackImageTTL
		}
		cacheControl = fmt.Sprintf("max-age=%d, public", ttl)
		expires = time.Now().Add(time.Second * time.Duration(ttl)).Format(http.TimeFormat)
	}

	if len(cacheControl) > 0 {
		rw.Header().Set("Cache-Control", cacheControl)
	}
	if len(expires) > 0 {
		rw.Header().Set("Expires", expires)
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

	if po.Dpr != 1 {
		rw.Header().Set("Content-DPR", strconv.FormatFloat(po.Dpr, 'f', 2, 32))
	}

	setCacheControl(rw, originData.Headers)
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
	rw.Write(resultData.Data)

	router.LogResponse(
		reqID, r, statusCode, nil,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)
}

func respondWithNotModified(reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string, originHeaders map[string]string) {
	setCacheControl(rw, originHeaders)
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

func sendErrAndPanic(ctx context.Context, errType string, err error) {
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

	if queueSem != nil {
		token, aquired := queueSem.TryAquire()
		if !aquired {
			panic(ierrors.New(429, "Too many requests", "Too many requests"))
		}
		defer token.Release()
	}

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

	if !security.VerifySourceURL(imageURL) {
		sendErrAndPanic(ctx, "security", ierrors.New(
			404,
			fmt.Sprintf("Source URL is not allowed: %s", imageURL),
			"Invalid source",
		))
	}

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

	// The heavy part start here, so we need to restrict concurrency
	var processingSemToken *semaphore.Token
	func() {
		defer metrics.StartQueueSegment(ctx)()

		var aquired bool
		processingSemToken, aquired = processingSem.Aquire(ctx)
		if !aquired {
			// We don't actually need to check timeout here,
			// but it's an easy way to check if this is an actual timeout
			// or the request was cancelled
			checkErr(ctx, "queue", router.CheckTimeout(ctx))
		}
	}()
	defer processingSemToken.Release()

	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	statusCode := http.StatusOK

	originData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartDownloadingSegment(ctx)()

		var cookieJar *cookiejar.Jar

		if config.CookiePassthrough {
			cookieJar, err = cookies.JarFromRequest(r)
			checkErr(ctx, "download", err)
		}

		return imagedata.Download(imageURL, "source image", imgRequestHeader, cookieJar)
	}()

	if err == nil {
		defer originData.Close()
	} else if nmErr, ok := err.(*imagedata.ErrorNotModified); ok && config.ETagEnabled {
		rw.Header().Set("ETag", etagHandler.GenerateExpectedETag())
		respondWithNotModified(reqID, r, rw, po, imageURL, nmErr.Headers)
		return
	} else {
		ierr, ierrok := err.(*ierrors.Error)
		if ierrok {
			statusCode = ierr.StatusCode
		}
		if config.ReportDownloadingErrors && (!ierrok || ierr.Unexpected) {
			errorreport.Report(err, r)
		}

		metrics.SendError(ctx, "download", err)

		if imagedata.FallbackImage == nil {
			panic(err)
		}

		log.Warningf("Could not load image %s. Using fallback image. %s", imageURL, err.Error())
		if config.FallbackImageHTTPCode > 0 {
			statusCode = config.FallbackImageHTTPCode
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

	if originData.Type == po.Format || po.Format == imagetype.Unknown {
		// Don't process SVG
		if originData.Type == imagetype.SVG {
			if config.SanitizeSvg {
				sanitized, svgErr := svg.Satitize(originData)
				checkErr(ctx, "svg_processing", svgErr)

				// Since we'll replace origin data, it's better to close it to return
				// it's buffer to the pool
				originData.Close()

				originData = sanitized
			}

			respondWithImage(reqID, r, rw, statusCode, originData, po, imageURL, originData)
			return
		}

		if len(po.SkipProcessingFormats) > 0 {
			for _, f := range po.SkipProcessingFormats {
				if f == originData.Type {
					respondWithImage(reqID, r, rw, statusCode, originData, po, imageURL, originData)
					return
				}
			}
		}
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
