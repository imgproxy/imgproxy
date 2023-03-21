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
	responseGzipBufPool *bufPool
	responseGzipPool    *gzipPool

	processingSem chan struct{}

	headerVaryValue string
	fallbackImage   *imageData
)

func initProcessingHandler() error {
	var err error

	processingSem = make(chan struct{}, conf.Concurrency)

	if conf.GZipCompression > 0 {
		responseGzipBufPool = newBufPool("gzip", conf.Concurrency, conf.GZipBufferSize)
		if responseGzipPool, err = newGzipPool(conf.Concurrency); err != nil {
			return err
		}
	}

	vary := make([]string, 0)

	if conf.EnableWebpDetection || conf.EnforceWebp {
		vary = append(vary, "Accept")
	}

	if conf.GZipCompression > 0 {
		vary = append(vary, "Accept-Encoding")
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

	if conf.CacheControlPassthrough {
		cacheControl = getCacheControlHeader(ctx)
		expires = getExpiresHeader(ctx)
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
		imgdata := getImageData(ctx)
		rw.Header().Set("X-Origin-Content-Length", strconv.Itoa(len(imgdata.Data)))
	}

	if conf.GZipCompression > 0 && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		buf := responseGzipBufPool.Get(0)
		defer responseGzipBufPool.Put(buf)

		gz := responseGzipPool.Get(buf)
		defer responseGzipPool.Put(gz)

		gz.Write(data)
		gz.Close()

		rw.Header().Set("Content-Encoding", "gzip")
		rw.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

		rw.WriteHeader(200)
		rw.Write(buf.Bytes())
	} else {
		rw.Header().Set("Content-Length", strconv.Itoa(len(data)))
		rw.WriteHeader(200)
		rw.Write(data)
	}

	imageURL := getImageURL(ctx)

	logResponse(reqID, r, 200, nil, &imageURL, po)
	// logResponse(reqID, r, 200, getTimerSince(ctx), getImageURL(ctx), po))
}

func respondWithNotModified(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter) {
	rw.WriteHeader(304)

	imageURL := getImageURL(ctx)

	logResponse(reqID, r, 304, nil, &imageURL, getProcessingOptions(ctx))
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cachePath string
	r, cachePath = beforeProcessing(r)

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

	ctx, downloadcancel, err := downloadImage(ctx)
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

		logWarning("Could not load image %s. Using fallback image. %s", getImageURL(ctx), err.Error())
		ctx = context.WithValue(ctx, imageDataCtxKey, fallbackImage)
	}

	checkTimeout(ctx)

	if conf.ETagEnabled {
		eTag := calcETag(ctx)
		rw.Header().Set("ETag", eTag)

		if eTag == r.Header.Get("If-None-Match") {
			respondWithNotModified(ctx, reqID, r, rw)
			return
		}
	}

	checkTimeout(ctx)

	if len(conf.SkipProcessingFormats) > 0 {
		imgdata := getImageData(ctx)
		po := getProcessingOptions(ctx)

		if imgdata.Type == po.Format || po.Format == imageTypeUnknown {
			for _, f := range conf.SkipProcessingFormats {
				if f == imgdata.Type {
					po.Format = imgdata.Type
					respondWithImage(ctx, reqID, r, rw, imgdata.Data)
					return
				}
			}
		}
	}

	imgData, processcancel, err := processImage(ctx)
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

	// copy imgData for thread safety
	imgDataForS3 := make([]byte, len(imgData))
	copy(imgDataForS3, imgData)
	uploaded := beforeResponse(imgDataForS3, cachePath)

	checkTimeout(ctx)

	respondWithImage(ctx, reqID, r, rw, imgData)

	// Waiting for S3 Upload to finish.
	<-uploaded
}
