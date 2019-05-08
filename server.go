package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid"
	"golang.org/x/net/netutil"
)

const (
	healthPath                         = "/health"
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

	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidMethod = newError(422, "Invalid request method", "Method doesn't allowed")
	errInvalidSecret = newError(403, "Invalid secret", "Forbidden")

	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)

	responseGzipBufPool *bufPool
	responseGzipPool    *gzipPool
)

type httpHandler struct {
	sem chan struct{}
}

func newHTTPHandler() *httpHandler {
	return &httpHandler{make(chan struct{}, conf.Concurrency)}
}

func startServer() *http.Server {
	l, err := net.Listen("tcp", conf.Bind)
	if err != nil {
		logFatal(err.Error())
	}
	s := &http.Server{
		Handler:        newHTTPHandler(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if conf.GZipCompression > 0 {
		responseGzipBufPool = newBufPool("gzip", conf.Concurrency, conf.GZipBufferSize)
		responseGzipPool = newGzipPool(conf.Concurrency)
	}

	go func() {
		logNotice("Starting server at %s", conf.Bind)
		if err := s.Serve(netutil.LimitListener(l, conf.MaxClients)); err != nil && err != http.ErrServerClosed {
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

func writeCORS(rw http.ResponseWriter) {
	if len(conf.AllowOrigin) > 0 {
		rw.Header().Set("Access-Control-Allow-Origin", conf.AllowOrigin)
		rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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

func respondWithOptions(reqID string, rw http.ResponseWriter) {
	logResponse(reqID, 200, "Respond with options")
	rw.WriteHeader(200)
}

func respondWithNotModified(reqID string, rw http.ResponseWriter) {
	logResponse(reqID, 200, "Not modified")
	rw.WriteHeader(304)
}

func generateRequestID(rw http.ResponseWriter, r *http.Request) (reqID string) {
	reqID = r.Header.Get(xRequestIDHeader)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		reqID, _ = nanoid.Nanoid()
	}

	rw.Header().Set(xRequestIDHeader, reqID)

	return
}

func prepareAuthHeaderMust() []byte {
	if len(authHeaderMust) == 0 {
		authHeaderMust = []byte(fmt.Sprintf("Bearer %s", conf.Secret))
	}

	return authHeaderMust
}

func checkSecret(r *http.Request) bool {
	if len(conf.Secret) == 0 {
		return true
	}

	return subtle.ConstantTimeCompare(
		[]byte(r.Header.Get("Authorization")),
		prepareAuthHeaderMust(),
	) == 1
}

func (h *httpHandler) lock() {
	h.sem <- struct{}{}
}

func (h *httpHandler) unlock() {
	<-h.sem
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Server", "imgproxy")

	reqID := generateRequestID(rw, r)

	defer func() {
		if rerr := recover(); rerr != nil {
			if err, ok := rerr.(error); ok {
				reportError(err, r)

				if ierr, ok := err.(*imgproxyError); ok {
					respondWithError(reqID, rw, ierr)
				} else {
					respondWithError(reqID, rw, newUnexpectedError(err.Error(), 3))
				}
			} else {
				panic(rerr)
			}
		}
	}()

	logRequest(reqID, r)

	writeCORS(rw)

	if r.Method == http.MethodOptions {
		respondWithOptions(reqID, rw)
		return
	}

	if r.Method != http.MethodGet {
		panic(errInvalidMethod)
	}

	if r.URL.RequestURI() == healthPath {
		rw.WriteHeader(200)
		rw.Write(imgproxyIsRunningMsg)
		return
	}

	if !checkSecret(r) {
		panic(errInvalidSecret)
	}

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

	h.lock()
	defer h.unlock()

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
			respondWithNotModified(reqID, rw)
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
