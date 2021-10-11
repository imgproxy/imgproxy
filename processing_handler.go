package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	processingSem chan struct{}

	headerVaryValue string
)

type fallbackImageUsedCtxKey struct{}

func initProcessingHandler() {
	processingSem = make(chan struct{}, config.Concurrency)

	vary := make([]string, 0)

	if config.EnableWebpDetection || config.EnforceWebp {
		vary = append(vary, "Accept")
	}

	if config.EnableClientHints {
		vary = append(vary, "DPR", "Viewport-Width", "Width")
	}

	headerVaryValue = strings.Join(vary, ", ")
}

func respondWithImage(reqID string, r *http.Request, rw http.ResponseWriter, resultData *imagedata.ImageData, po *options.ProcessingOptions, originURL string, originData *imagedata.ImageData) {
	var contentDisposition string
	if len(po.Filename) > 0 {
		contentDisposition = resultData.Type.ContentDisposition(po.Filename)
	} else {
		contentDisposition = resultData.Type.ContentDispositionFromURL(originURL)
	}

	rw.Header().Set("Content-Type", resultData.Type.Mime())
	rw.Header().Set("Content-Disposition", contentDisposition)

	if po.Dpr != 1 {
		rw.Header().Set("Content-DPR", strconv.FormatFloat(po.Dpr, 'f', 2, 32))
	}

	if config.SetCanonicalHeader {
		if strings.HasPrefix(originURL, "https://") || strings.HasPrefix(originURL, "http://") {
			linkHeader := fmt.Sprintf(`<%s>; rel="canonical"`, originURL)
			rw.Header().Set("Link", linkHeader)
		}
	}

	var cacheControl, expires string

	if config.CacheControlPassthrough && originData.Headers != nil {
		if val, ok := originData.Headers["Cache-Control"]; ok {
			cacheControl = val
		}
		if val, ok := originData.Headers["Expires"]; ok {
			expires = val
		}
	}

	if len(cacheControl) == 0 && len(expires) == 0 {
		cacheControl = fmt.Sprintf("max-age=%d, public", config.TTL)
		expires = time.Now().Add(time.Second * time.Duration(config.TTL)).Format(http.TimeFormat)
	}

	if len(cacheControl) > 0 {
		rw.Header().Set("Cache-Control", cacheControl)
	}
	if len(expires) > 0 {
		rw.Header().Set("Expires", expires)
	}

	if len(headerVaryValue) > 0 {
		rw.Header().Set("Vary", headerVaryValue)
	}

	if config.EnableDebugHeaders {
		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(originData.Data)))
	}

	rw.Header().Set("Content-Length", strconv.Itoa(len(resultData.Data)))
	statusCode := 200
	if getFallbackImageUsed(r.Context()) {
		statusCode = config.FallbackImageHTTPCode
	}
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

func respondWithNotModified(reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, originURL string) {
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
	ctx, timeoutCancel := context.WithTimeout(r.Context(), time.Duration(config.WriteTimeout)*time.Second)
	defer timeoutCancel()

	var metricsCancel context.CancelFunc
	ctx, metricsCancel, rw = metrics.StartRequest(ctx, rw, r)
	defer metricsCancel()

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
		panic(ierrors.New(404, fmt.Sprintf("Invalid path: %s", path), "Invalid URL"))
	}

	if err := security.VerifySignature(signature, path); err != nil {
		panic(ierrors.New(403, err.Error(), "Forbidden"))
	}

	po, imageURL, err := options.ParsePath(path, r.Header)
	if err != nil {
		panic(err)
	}

	if !security.VerifySourceURL(imageURL) {
		panic(ierrors.New(404, fmt.Sprintf("Source URL is not allowed: %s", imageURL), "Invalid source"))
	}

	// SVG is a special case. Though saving to svg is not supported, SVG->SVG is.
	if !vips.SupportsSave(po.Format) && po.Format != imagetype.Unknown && po.Format != imagetype.SVG {
		panic(ierrors.New(
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
	select {
	case processingSem <- struct{}{}:
	case <-ctx.Done():
		// We don't actually need to check timeout here,
		// but it's an easy way to check if this is an actual timeout
		// or the request was cancelled
		router.CheckTimeout(ctx)
	}
	defer func() { <-processingSem }()

	originData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartDownloadingSegment(ctx)()
		return imagedata.Download(imageURL, "source image", imgRequestHeader)
	}()
	switch {
	case err == nil:
		defer originData.Close()
	case ierrors.StatusCode(err) == http.StatusNotModified:
		rw.Header().Set("ETag", etagHandler.GenerateExpectedETag())
		respondWithNotModified(reqID, r, rw, po, imageURL)
		return
	default:
		if ierr, ok := err.(*ierrors.Error); !ok || ierr.Unexpected {
			errorreport.Report(err, r)
		}

		metrics.SendError(ctx, "download", err)

		if imagedata.FallbackImage == nil {
			panic(err)
		}

		log.Warningf("Could not load image %s. Using fallback image. %s", imageURL, err.Error())
		r = r.WithContext(setFallbackImageUsedCtx(r.Context()))
		originData = imagedata.FallbackImage
	}

	router.CheckTimeout(ctx)

	if config.ETagEnabled && !getFallbackImageUsed(ctx) {
		imgDataMatch := etagHandler.SetActualImageData(originData)

		rw.Header().Set("ETag", etagHandler.GenerateActualETag())

		if imgDataMatch && etagHandler.ProcessingOptionsMatch() {
			respondWithNotModified(reqID, r, rw, po, imageURL)
			return
		}
	}

	router.CheckTimeout(ctx)

	if originData.Type == po.Format || po.Format == imagetype.Unknown {
		// Don't process SVG
		if originData.Type == imagetype.SVG {
			respondWithImage(reqID, r, rw, originData, po, imageURL, originData)
			return
		}

		if len(po.SkipProcessingFormats) > 0 {
			for _, f := range po.SkipProcessingFormats {
				if f == originData.Type {
					respondWithImage(reqID, r, rw, originData, po, imageURL, originData)
					return
				}
			}
		}
	}

	if !vips.SupportsLoad(originData.Type) {
		panic(ierrors.New(
			422,
			fmt.Sprintf("Source image format is not supported: %s", originData.Type),
			"Invalid URL",
		))
	}

	// At this point we can't allow requested format to be SVG as we can't save SVGs
	if po.Format == imagetype.SVG {
		panic(ierrors.New(422, "Resulting image format is not supported: svg", "Invalid URL"))
	}

	resultData, err := func() (*imagedata.ImageData, error) {
		defer metrics.StartProcessingSegment(ctx)()
		return processing.ProcessImage(ctx, originData, po)
	}()
	if err != nil {
		metrics.SendError(ctx, "processing", err)
		panic(err)
	}
	defer resultData.Close()

	router.CheckTimeout(ctx)

	respondWithImage(reqID, r, rw, resultData, po, imageURL, originData)
}

func setFallbackImageUsedCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, fallbackImageUsedCtxKey{}, true)
}

func getFallbackImageUsed(ctx context.Context) bool {
	result, _ := ctx.Value(fallbackImageUsedCtxKey{}).(bool)
	return result
}
