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
)

func initProcessingHandler() {
	processingSem = make(chan struct{}, conf.Concurrency)

	if conf.GZipCompression > 0 {
		responseGzipBufPool = newBufPool("gzip", conf.Concurrency, conf.GZipBufferSize)
		responseGzipPool = newGzipPool(conf.Concurrency)
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
}

func respondWithImage(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter, data []byte) {
	po := getProcessingOptions(ctx)

	var contentDisposition string
	if len(po.Filename) > 0 {
		contentDisposition = po.Format.ContentDisposition(po.Filename)
	} else {
		contentDisposition = po.Format.ContentDispositionFromURL(getImageURL(ctx))
	}

	rw.Header().Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))
	rw.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", conf.TTL))
	rw.Header().Set("Content-Type", po.Format.Mime())
	rw.Header().Set("Content-Disposition", contentDisposition)

	if len(headerVaryValue) > 0 {
		rw.Header().Set("Vary", headerVaryValue)
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

	if newRelicEnabled {
		var newRelicCancel context.CancelFunc
		ctx, newRelicCancel = startNewRelicTransaction(ctx, rw, r)
		defer newRelicCancel()
	}

	if prometheusEnabled {
		prometheusRequestsTotal.Inc()
		defer startPrometheusDuration(prometheusRequestDuration)()
	}

	processingSem <- struct{}{}
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
		panic(err)
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

	po := getProcessingOptions(ctx)
	var imageData []byte
	var processcancel context.CancelFunc
	if po.MaxBytes > 0 {
		imageData, processcancel, err = processImageMaxBytes(ctx)
	} else {
		imageData, processcancel, err = processImage(ctx)
	}
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
