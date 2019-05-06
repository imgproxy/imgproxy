package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid"
	"github.com/valyala/fasthttp"
)

const (
	contextDispositionFilenameFallback = "image"
	xRequestIDHeader                   = "X-Request-ID"
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

	authHeaderMust []byte

	healthPath = []byte("/health")

	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidMethod = newError(422, "Invalid request method", "Method doesn't allowed")
	errInvalidSecret = newError(403, "Invalid secret", "Forbidden")

	responseGzipPool *gzipPool

	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

type httpHandler struct {
	sem chan struct{}
}

func newHTTPHandler() *httpHandler {
	return &httpHandler{make(chan struct{}, conf.Concurrency)}
}

func startServer() *fasthttp.Server {
	handler := newHTTPHandler()

	server := &fasthttp.Server{
		Name:        "imgproxy",
		Handler:     handler.ServeHTTP,
		Concurrency: conf.MaxClients,
		ReadTimeout: time.Duration(conf.ReadTimeout) * time.Second,
	}

	if conf.GZipCompression > 0 {
		responseGzipPool = newGzipPool(conf.Concurrency)
	}

	if conf.ETagEnabled {
		eTagCalcPool = newEtagPool(conf.Concurrency)
	}

	go func() {
		logNotice("Starting server at %s", conf.Bind)
		if err := server.ListenAndServe(conf.Bind); err != nil {
			logFatal(err.Error())
		}
	}()

	return server
}

func shutdownServer(s *fasthttp.Server) {
	logNotice("Shutting down the server...")
	s.Shutdown()
}

func writeCORS(rctx *fasthttp.RequestCtx) {
	if len(conf.AllowOrigin) > 0 {
		rctx.Response.Header.Set("Access-Control-Allow-Origin", conf.AllowOrigin)
		rctx.Response.Header.Set("Access-Control-Allow-Methods", "GET, OPTIONs")
	}
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

func respondWithImage(ctx context.Context, reqID string, rctx *fasthttp.RequestCtx, data []byte) {
	po := getProcessingOptions(ctx)

	rctx.SetStatusCode(200)

	rctx.Response.Header.Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))
	rctx.Response.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d, public", conf.TTL))
	rctx.Response.Header.Set("Content-Type", mimes[po.Format])
	rctx.Response.Header.Set("Content-Disposition", contentDisposition(getImageURL(ctx), po.Format))

	addVaryHeader(rctx)

	if conf.GZipCompression > 0 && rctx.Request.Header.HasAcceptEncoding("gzip") {
		gz := responseGzipPool.Get(rctx)
		defer responseGzipPool.Put(gz)

		gz.Write(data)
		gz.Close()

		rctx.Response.Header.Set("Content-Encoding", "gzip")
	} else {
		rctx.SetBody(data)
	}

	logResponse(reqID, 200, fmt.Sprintf("Processed in %s: %s; %+v", getTimerSince(ctx), getImageURL(ctx), po))
}

func addVaryHeader(rctx *fasthttp.RequestCtx) {
	vary := make([]string, 0, 5)

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
		rctx.Response.Header.Set("Vary", strings.Join(vary, ", "))
	}
}

func respondWithError(reqID string, rctx *fasthttp.RequestCtx, err *imgproxyError) {
	logResponse(reqID, err.StatusCode, err.Message)

	rctx.SetStatusCode(err.StatusCode)
	rctx.SetBodyString(err.PublicMessage)
}

func respondWithOptions(reqID string, rctx *fasthttp.RequestCtx) {
	logResponse(reqID, 200, "Respond with options")
	rctx.SetStatusCode(200)
}

func respondWithNotModified(reqID string, rctx *fasthttp.RequestCtx) {
	logResponse(reqID, 304, "Not modified")
	rctx.SetStatusCode(304)
}

func generateRequestID(rctx *fasthttp.RequestCtx) (reqID string) {
	reqIDb := rctx.Request.Header.Peek(xRequestIDHeader)

	if len(reqIDb) > 0 && requestIDRe.Match(reqIDb) {
		reqID = string(reqIDb)
	} else {
		reqID, _ = nanoid.Nanoid()
	}

	rctx.Response.Header.Set(xRequestIDHeader, reqID)

	return
}

func prepareAuthHeaderMust() []byte {
	if len(authHeaderMust) == 0 {
		authHeaderMust = []byte(fmt.Sprintf("Bearer %s", conf.Secret))
	}

	return authHeaderMust
}

func checkSecret(rctx *fasthttp.RequestCtx) bool {
	if len(conf.Secret) == 0 {
		return true
	}

	return subtle.ConstantTimeCompare(
		rctx.Request.Header.Peek("Authorization"),
		prepareAuthHeaderMust(),
	) == 1
}

func requestCtxToRequest(rctx *fasthttp.RequestCtx) *http.Request {
	if r, ok := rctx.UserValue("httpRequest").(*http.Request); ok {
		return r
	}

	reqURL, _ := url.Parse(rctx.Request.URI().String())

	r := &http.Request{
		Method:     http.MethodGet, // Only GET is supported
		URL:        reqURL,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     make(http.Header),
		Body:       http.NoBody,
		Host:       reqURL.Host,
		RequestURI: reqURL.RequestURI(),
		RemoteAddr: rctx.RemoteAddr().String(),
	}

	rctx.Request.Header.VisitAll(func(key, value []byte) {
		r.Header.Add(string(key), string(value))
	})

	rctx.SetUserValue("httpRequest", r)

	return r
}

func (h *httpHandler) lock() {
	h.sem <- struct{}{}
}

func (h *httpHandler) unlock() {
	<-h.sem
}

func (h *httpHandler) ServeHTTP(rctx *fasthttp.RequestCtx) {
	reqID := generateRequestID(rctx)

	defer func() {
		if rerr := recover(); rerr != nil {
			if err, ok := rerr.(error); ok {
				reportError(err, requestCtxToRequest(rctx))

				if ierr, ok := err.(*imgproxyError); ok {
					respondWithError(reqID, rctx, ierr)
				} else {
					respondWithError(reqID, rctx, newUnexpectedError(err, 4))
				}
			} else {
				panic(rerr)
			}
		}
	}()

	logRequest(reqID, rctx)

	writeCORS(rctx)

	if rctx.Request.Header.IsOptions() {
		respondWithOptions(reqID, rctx)
		return
	}

	if !rctx.Request.Header.IsGet() {
		panic(errInvalidMethod)
	}

	if bytes.Equal(rctx.RequestURI(), healthPath) {
		rctx.SetStatusCode(200)
		rctx.SetBody(imgproxyIsRunningMsg)
		return
	}

	if !checkSecret(rctx) {
		panic(errInvalidSecret)
	}

	ctx := context.Background()

	if newRelicEnabled {
		var newRelicCancel context.CancelFunc
		ctx, newRelicCancel = startNewRelicTransaction(ctx, requestCtxToRequest(rctx))
		defer newRelicCancel()
	}

	if prometheusEnabled {
		prometheusRequestsTotal.Inc()
		defer startPrometheusDuration(prometheusRequestDuration)()
	}

	h.lock()
	defer h.unlock()

	ctx, timeoutCancel := startTimer(ctx, time.Duration(conf.WriteTimeout)*time.Second)
	defer timeoutCancel()

	ctx, err := parsePath(ctx, rctx)
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
		eTag, etagcancel := calcETag(ctx)
		defer etagcancel()

		rctx.Response.Header.SetBytesV("ETag", eTag)

		if bytes.Equal(eTag, rctx.Request.Header.Peek("If-None-Match")) {
			respondWithNotModified(reqID, rctx)
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

	respondWithImage(ctx, reqID, rctx, imageData)
}
