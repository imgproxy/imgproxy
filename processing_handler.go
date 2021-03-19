package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	processingSem chan struct{}

	headerVaryValue string
	fallbackImage   *imageData
)

const (
	fallbackImageUsedCtxKey = ctxKey("fallbackImageUsed")
)

func initProcessingHandler() error {
	var err error

	processingSem = make(chan struct{}, conf.Concurrency)

	vary := make([]string, 0)

	if conf.EnableWebpDetection || conf.EnforceWebp {
		vary = append(vary, "Accept")
	}

	if conf.EnableClientHints {
		vary = append(vary, "DPR", "Viewport-Width", "Width")
	}

	headerVaryValue = strings.Join(vary, ", ")

	if fallbackImage, err = getFallbackImageData(); err != nil {
		return err
	}

	return nil
}

func respondWithImage(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter, data []byte) {
	po := getProcessingOptions(ctx)
	imgdata := getImageData(ctx)

	var contentDisposition string
	if len(po.Filename) > 0 {
		contentDisposition = po.Format.ContentDisposition(po.Filename)
	} else {
		contentDisposition = po.Format.ContentDispositionFromURL(getImageURL(ctx))
	}

	rw.Header().Set("Content-Type", po.Format.Mime())
	rw.Header().Set("Content-Disposition", contentDisposition)

	if conf.SetCanonicalHeader {
		origin := getImageURL(ctx)
		if strings.HasPrefix(origin, "https://") || strings.HasPrefix(origin, "http://") {
			linkHeader := fmt.Sprintf(`<%s>; rel="canonical"`, origin)
			rw.Header().Set("Link", linkHeader)
		}
	}

	var cacheControl, expires string

	if conf.CacheControlPassthrough && imgdata.Headers != nil {
		if val, ok := imgdata.Headers["Cache-Control"]; ok {
			cacheControl = val
		}
		if val, ok := imgdata.Headers["Expires"]; ok {
			expires = val
		}
	}

	if len(cacheControl) == 0 && len(expires) == 0 {
		cacheControl = fmt.Sprintf("max-age=%d, public", conf.TTL)
		expires = time.Now().Add(time.Second * time.Duration(conf.TTL)).Format(http.TimeFormat)
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

	if conf.EnableDebugHeaders {
		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(imgdata.Data)))
	}

	rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
	statusCode := 200
	if getFallbackImageUsed(ctx) {
		statusCode = conf.FallbackImageHTTPCode
	}
	rw.WriteHeader(statusCode)
	rw.Write(data)

	imageURL := getImageURL(ctx)

	logResponse(reqID, r, statusCode, nil, &imageURL, po)
}

func respondWithNotModified(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter) {
	rw.WriteHeader(304)

	imageURL := getImageURL(ctx)

	logResponse(reqID, r, 304, nil, &imageURL, getProcessingOptions(ctx))
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if newRelicEnabled {
		var newRelicCancel context.CancelFunc
		ctx, newRelicCancel, rw = startNewRelicTransaction(ctx, rw, r)
		defer newRelicCancel()
	}

	if prometheusEnabled {
		prometheusRequestsTotal.Inc()
		defer startPrometheusDuration(prometheusRequestDuration)()
	}

	select {
	case processingSem <- struct{}{}:
	case <-ctx.Done():
		panic(newError(499, "Request was cancelled before processing", "Cancelled"))
	}
	defer func() { <-processingSem }()

	ctx, timeoutCancel := context.WithTimeout(ctx, time.Duration(conf.WriteTimeout)*time.Second)
	defer timeoutCancel()

	ctx, err := parsePath(ctx, r)
	if err != nil {
		panic(err)
	}

	ctx, downloadcancel, err := downloadImageCtx(ctx)
	defer downloadcancel()
	if err != nil {
		if newRelicEnabled {
			sendErrorToNewRelic(ctx, err)
		}
		if prometheusEnabled {
			incrementPrometheusErrorsTotal("download")
		}

		if fallbackImage == nil {
			panic(err)
		}

		if ierr, ok := err.(*imgproxyError); !ok || ierr.Unexpected {
			reportError(err, r)
		}

		logWarning("Could not load image. Using fallback image: %s", err.Error())
		ctx = setFallbackImageUsedCtx(ctx)
		ctx = context.WithValue(ctx, imageDataCtxKey, fallbackImage)
	}

	checkTimeout(ctx)

	if conf.ETagEnabled && !getFallbackImageUsed(ctx) {
		eTag := calcETag(ctx)
		rw.Header().Set("ETag", eTag)

		if eTag == r.Header.Get("If-None-Match") {
			respondWithNotModified(ctx, reqID, r, rw)
			return
		}
	}

	checkTimeout(ctx)

	po := getProcessingOptions(ctx)
	if len(conf.SkipProcessingFormats) > 0 || len(po.SkipProcessingFormats) > 0 {
		imgdata := getImageData(ctx)

		if imgdata.Type == po.Format || po.Format == imageTypeUnknown {
			for _, f := range append(conf.SkipProcessingFormats, po.SkipProcessingFormats...) {
				if f == imgdata.Type {
					po.Format = imgdata.Type
					respondWithImage(ctx, reqID, r, rw, imgdata.Data)
					return
				}
			}
		}
	}

	imageData, processcancel, err := processImage(ctx)
	defer processcancel()
	if err != nil {
		if newRelicEnabled {
			sendErrorToNewRelic(ctx, err)
		}
		if prometheusEnabled {
			incrementPrometheusErrorsTotal("processing")
		}
		panic(err)
	}

	checkTimeout(ctx)

	respondWithImage(ctx, reqID, r, rw, imageData)
}

func setFallbackImageUsedCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, fallbackImageUsedCtxKey, true)
}

func getFallbackImageUsed(ctx context.Context) bool {
	result, _ := ctx.Value(fallbackImageUsedCtxKey).(bool)
	return result
}
