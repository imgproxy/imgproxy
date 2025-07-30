package main

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/handlererr"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagestreamer"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/stemext"
	"github.com/imgproxy/imgproxy/v3/svg"
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

func respondWithImage(hw *headerwriter.Writer, reqID string, r *http.Request, rw http.ResponseWriter, statusCode int, resultData *imagedata.ImageData, po *options.ProcessingOptions, originURL string, originData *imagedata.ImageData) {
	url, err := url.Parse(originURL)
	handlererr.Check(r.Context(), handlererr.ErrTypePathParsing, err)

	stem, ext := stemext.FromURL(url).
		OverrideStem(po.Filename).
		OverrideExt(resultData.Type.Ext()).
		StemExtWithFallback()

	hw.SetMaxAgeFromExpires(po.Expires)
	hw.SetContentDisposition(stem, ext, po.ReturnAttachment)
	hw.SetContentType(resultData.Type.Mime())
	hw.SetLastModified()
	hw.SetVary()

	// TODO: think about moving this to the headerwriter
	if config.EnableDebugHeaders {
		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(originData.Data)))
		rw.Header().Set("X-Origin-Width", resultData.Headers["X-Origin-Width"])
		rw.Header().Set("X-Origin-Height", resultData.Headers["X-Origin-Height"])
		rw.Header().Set("X-Result-Width", resultData.Headers["X-Result-Width"])
		rw.Header().Set("X-Result-Height", resultData.Headers["X-Result-Height"])
	}

	hw.SetContentLength(len(resultData.Data))
	hw.SetCanonical()
	hw.Write(rw)

	rw.WriteHeader(statusCode)

	_, err = rw.Write(resultData.Data)

	var ierr *ierrors.Error
	if err != nil {
		ierr = newResponseWriteError(err)

		if config.ReportIOErrors {
			handlererr.Send(r.Context(), handlererr.ErrTypeIO, ierr)
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

func respondWithNotModified(hw *headerwriter.Writer, reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string, originHeaders map[string]string) {
	hw.SetMaxAgeFromExpires(po.Expires)
	hw.SetVary()
	hw.Write(rw)

	rw.WriteHeader(304)
	router.LogResponse(
		reqID, r, 304, nil,
		log.Fields{
			"image_url":          originURL,
			"processing_options": po,
		},
	)
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) {
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
		handlererr.SendAndPanic(ctx, handlererr.ErrTypePathParsing, newInvalidURLErrorf(
			http.StatusNotFound, "Invalid path: %s", path),
		)
	}

	path = fixPath(path)

	if err := security.VerifySignature(signature, path); err != nil {
		handlererr.SendAndPanic(ctx, handlererr.ErrTypeSecurity, err)
	}

	po, imageURL, err := options.ParsePath(path, r.Header)
	handlererr.Check(ctx, handlererr.ErrTypePathParsing, err)

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
	handlererr.Check(ctx, handlererr.ErrTypeSecurity, err)

	if po.Raw {
		sf := imagestreamer.NewService(
			imagestreamer.NewConfigFromEnv(),
			imagedata.Fetcher,
			headerwriter.NewFactory(headerwriter.NewConfigFromEnv()),
		)

		p := imagestreamer.Request{
			UserRequest:       r,
			ImageURL:          imageURL,
			ReqID:             reqID,
			ProcessingOptions: po,
		}

		sf.Stream(ctx, &p, rw)
		return
	}

	// SVG is a special case. Though saving to svg is not supported, SVG->SVG is.
	if !vips.SupportsSave(po.Format) && po.Format != imagetype.Unknown && po.Format != imagetype.SVG {
		handlererr.SendAndPanic(ctx, handlererr.ErrTypePathParsing, newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Resulting image format is not supported: %s", po.Format,
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

	// ???
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
	func() {
		defer metrics.StartQueueSegment(ctx)()

		err = processingSem.Acquire(ctx, 1)
		if err != nil {
			// We don't actually need to check timeout here,
			// but it's an easy way to check if this is an actual timeout
			// or the request was canceled
			handlererr.Check(ctx, handlererr.ErrTypeQueue, router.CheckTimeout(ctx))
			// We should never reach this line as err could be only ctx.Err()
			// and we've already checked for it. But beter safe than sorry
			handlererr.SendAndPanic(ctx, handlererr.ErrTypeQueue, err)
		}
	}()
	defer processingSem.Release(1)

	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	statusCode := http.StatusOK

	originData, originResponseHeaders, err := func() (*imagedata.ImageData, http.Header, error) {
		defer metrics.StartDownloadingSegment(ctx, metrics.Meta{
			metrics.MetaSourceImageURL:    metricsMeta[metrics.MetaSourceImageURL],
			metrics.MetaSourceImageOrigin: metricsMeta[metrics.MetaSourceImageOrigin],
		})()

		downloadOpts := imagedata.DownloadOptions{
			Header:    imgRequestHeader,
			CookieJar: nil,
		}

		if config.CookiePassthrough {
			downloadOpts.CookieJar, err = cookies.JarFromRequest(r)
			handlererr.Check(ctx, handlererr.ErrTypeDownload, err)
		}

		return imagedata.Download(ctx, imageURL, "source image", downloadOpts, po.SecurityOptions)
	}()

	hwf := headerwriter.NewFactory(headerwriter.NewConfigFromEnv())
	hw := hwf.NewHeaderWriter(originResponseHeaders, imageURL)

	var nmErr imagefetcher.NotModifiedError

	switch {
	case err == nil:
		defer originData.Close()

	case errors.As(err, &nmErr):
		if config.ETagEnabled && len(etagHandler.ImageEtagExpected()) != 0 {
			rw.Header().Set("ETag", etagHandler.GenerateExpectedETag())
		}

		h := make(map[string]string)
		for k := range nmErr.Headers() {
			h[k] = nmErr.Headers().Get(k)
		}

		respondWithNotModified(hw, reqID, r, rw, po, imageURL, h)
		return

	default:
		// This may be a request timeout error or a request cancelled error.
		// Check it before moving further
		handlererr.Check(ctx, handlererr.ErrTypeTimeout, router.CheckTimeout(ctx))

		ierr := ierrors.Wrap(err, 0)
		if config.ReportDownloadingErrors {
			ierr = ierrors.Wrap(ierr, 0, ierrors.WithShouldReport(true))
		}

		handlererr.Send(ctx, handlererr.ErrTypeDownload, ierr)

		if imagedata.FallbackImage == nil {
			panic(ierr)
		}

		// We didn't panic, so the error is not reported.
		// Report it now
		if ierr.ShouldReport() {
			errorreport.Report(ierr, r)
		}

		log.WithField("request_id", reqID).Warningf("Could not load image %s. Using fallback image. %s", imageURL, ierr.Error())

		if config.FallbackImageHTTPCode > 0 {
			statusCode = config.FallbackImageHTTPCode
		} else {
			statusCode = ierr.StatusCode()
		}

		hw.SetMaxAge(config.FallbackImageTTL)

		if config.FallbackImageTTL > 0 {
			hw.SetIsFallbackImage()
		}

		originData = imagedata.FallbackImage
	}

	handlererr.Check(ctx, handlererr.ErrTypeTimeout, router.CheckTimeout(ctx))

	if config.ETagEnabled && statusCode == http.StatusOK {
		imgDataMatch := etagHandler.SetActualImageData(originData)

		rw.Header().Set("ETag", etagHandler.GenerateActualETag())

		if imgDataMatch && etagHandler.ProcessingOptionsMatch() {
			respondWithNotModified(hw, reqID, r, rw, po, imageURL, originData.Headers)
			return
		}
	}

	handlererr.Check(ctx, handlererr.ErrTypeTimeout, router.CheckTimeout(ctx))

	// Skip processing svg with unknown or the same destination imageType
	// if it's not forced by AlwaysRasterizeSvg option
	// Also skip processing if the format is in SkipProcessingFormats
	shouldSkipProcessing := (originData.Type == po.Format || po.Format == imagetype.Unknown) &&
		(slices.Contains(po.SkipProcessingFormats, originData.Type) ||
			originData.Type == imagetype.SVG && !config.AlwaysRasterizeSvg)

	if shouldSkipProcessing {
		if originData.Type == imagetype.SVG && config.SanitizeSvg {
			sanitized, svgErr := svg.Sanitize(originData)
			handlererr.Check(ctx, handlererr.ErrTypeSvgProcessing, svgErr)

			defer sanitized.Close()

			respondWithImage(hw, reqID, r, rw, statusCode, sanitized, po, imageURL, originData)
			return
		}

		respondWithImage(hw, reqID, r, rw, statusCode, originData, po, imageURL, originData)
		return
	}

	if !vips.SupportsLoad(originData.Type) {
		handlererr.SendAndPanic(ctx, handlererr.ErrTypeProcessing, newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Source image format is not supported: %s", originData.Type,
		))
	}

	// At this point we can't allow requested format to be SVG as we can't save SVGs
	if po.Format == imagetype.SVG {
		handlererr.SendAndPanic(ctx, handlererr.ErrTypeProcessing, newInvalidURLErrorf(
			http.StatusUnprocessableEntity,
			"Resulting image format is not supported: svg",
		))
	}

	resultData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartProcessingSegment(ctx, metrics.Meta{
			metrics.MetaProcessingOptions: metricsMeta[metrics.MetaProcessingOptions],
		})()
		return processing.ProcessImage(ctx, originData, po)
	}()
	handlererr.Check(ctx, handlererr.ErrTypeProcessing, err)

	defer resultData.Close()

	handlererr.Check(ctx, handlererr.ErrTypeTimeout, router.CheckTimeout(ctx))

	respondWithImage(hw, reqID, r, rw, statusCode, resultData, po, imageURL, originData)
}
