package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/netutil"
)

const (
	contextDispositionFilenameFallback = "image"
)

var (
	mimes = map[imageType]string{
		imageTypeJPEG: "image/jpeg",
		imageTypePNG:  "image/png",
		imageTypeWEBP: "image/webp",
		imageTypeGIF:  "image/gif",
		imageTypeICO:  "image/x-icon",
	}

	contentDispositionsFmt = map[imageType]string{
		imageTypeJPEG: "inline; filename=\"%s.jpg\"",
		imageTypePNG:  "inline; filename=\"%s.png\"",
		imageTypeWEBP: "inline; filename=\"%s.webp\"",
		imageTypeGIF:  "inline; filename=\"%s.gif\"",
		imageTypeICO:  "inline; filename=\"%s.ico\"",
	}

	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidMethod = newError(422, "Invalid request method", "Method doesn't allowed")
	errInvalidSecret = newError(403, "Invalid secret", "Forbidden")

	responseGzipBufPool *bufPool
	responseGzipPool    *gzipPool

	processingSem chan struct{}
)

func buildRouter() *router {
	r := newRouter()

	r.PanicHandler = handlePanic

	r.GET("/health", handleHealth)
	r.GET("/", withCORS(withSecret(handleProcessing)))
	r.OPTIONS("/", withCORS(handleOptions))

	return r
}

func startServer() *http.Server {
	processingSem = make(chan struct{}, conf.Concurrency)

	l, err := net.Listen("tcp", conf.Bind)
	if err != nil {
		logFatal(err.Error())
	}
	l = netutil.LimitListener(l, conf.MaxClients)

	s := &http.Server{
		Handler:        buildRouter(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if conf.GZipCompression > 0 {
		responseGzipBufPool = newBufPool("gzip", conf.Concurrency, conf.GZipBufferSize)
		responseGzipPool = newGzipPool(conf.Concurrency)
	}

	go func() {
		logNotice("Starting server at %s", conf.Bind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			logFatal(err.Error())
		}
	}()

	return s
}

func shutdownServer(s *http.Server) {
	logNotice("Shutting down the server...")

	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	s.Shutdown(ctx)
}

func contentDisposition(imageURL string, imgtype imageType) string {
	url, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Sprintf(contentDispositionsFmt[imgtype], contextDispositionFilenameFallback)
	}

	_, filename := filepath.Split(url.Path)
	if len(filename) == 0 {
		return fmt.Sprintf(contentDispositionsFmt[imgtype], contextDispositionFilenameFallback)
	}

	return fmt.Sprintf(contentDispositionsFmt[imgtype], strings.TrimSuffix(filename, filepath.Ext(filename)))
}

func respondWithImage(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter, data []byte) {
	po := getProcessingOptions(ctx)

	rw.Header().Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))
	rw.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", conf.TTL))
	rw.Header().Set("Content-Type", mimes[po.Format])
	rw.Header().Set("Content-Disposition", contentDisposition(getImageURL(ctx), po.Format))

	addVaryHeader(rw)

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

	logResponse(reqID, 200, fmt.Sprintf("Processed in %s: %s; %+v", getTimerSince(ctx), getImageURL(ctx), po))
}

func addVaryHeader(rw http.ResponseWriter) {
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

	if len(vary) > 0 {
		rw.Header().Set("Vary", strings.Join(vary, ", "))
	}
}

func respondWithError(reqID string, rw http.ResponseWriter, err *imgproxyError) {
	logResponse(reqID, err.StatusCode, err.Message)

	rw.WriteHeader(err.StatusCode)

	if conf.DevelopmentErrorsMode {
		rw.Write([]byte(err.Message))
	} else {
		rw.Write([]byte(err.PublicMessage))
	}
}

func withCORS(h routeHandler) routeHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		if len(conf.AllowOrigin) > 0 {
			rw.Header().Set("Access-Control-Allow-Origin", conf.AllowOrigin)
			rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}

		h(reqID, rw, r)
	}
}

func withSecret(h routeHandler) routeHandler {
	if len(conf.Secret) == 0 {
		return h
	}

	authHeader := []byte(fmt.Sprintf("Bearer %s", conf.Secret))

	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), authHeader) == 1 {
			h(reqID, rw, r)
		} else {
			respondWithError(reqID, rw, errInvalidSecret)
		}
	}
}

func handlePanic(reqID string, rw http.ResponseWriter, r *http.Request, err error) {
	reportError(err, r)

	if ierr, ok := err.(*imgproxyError); ok {
		respondWithError(reqID, rw, ierr)
	} else {
		respondWithError(reqID, rw, newUnexpectedError(err.Error(), 3))
	}
}

func handleHealth(reqID string, rw http.ResponseWriter, r *http.Request) {
	logResponse(reqID, 200, string(imgproxyIsRunningMsg))
	rw.WriteHeader(200)
	rw.Write(imgproxyIsRunningMsg)
}

func handleOptions(reqID string, rw http.ResponseWriter, r *http.Request) {
	logResponse(reqID, 200, "Respond with options")
	rw.WriteHeader(200)
}

func handleProcessing(reqID string, rw http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

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

	ctx, timeoutCancel := startTimer(ctx, time.Duration(conf.WriteTimeout)*time.Second)
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
			logResponse(reqID, 304, "Not modified")
			rw.WriteHeader(304)
			return
		}
	}

	checkTimeout(ctx)

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
