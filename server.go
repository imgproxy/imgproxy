package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid"
	"golang.org/x/net/netutil"
)

var mimes = map[imageType]string{
	imageTypeJPEG: "image/jpeg",
	imageTypePNG:  "image/png",
	imageTypeWEBP: "image/webp",
}

type httpHandler struct {
	sem chan struct{}
}

func newHTTPHandler() *httpHandler {
	return &httpHandler{make(chan struct{}, conf.Concurrency)}
}

func startServer() *http.Server {
	l, err := net.Listen("tcp", conf.Bind)
	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Handler:        newHTTPHandler(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("Starting server at %s\n", conf.Bind)
		log.Fatal(s.Serve(netutil.LimitListener(l, conf.MaxClients)))
	}()

	return s
}

func shutdownServer(s *http.Server) {
	log.Println("Shutting down the server...")

	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	s.Shutdown(ctx)
}

func logResponse(status int, msg string) {
	var color int

	if status >= 500 {
		color = 31
	} else if status >= 400 {
		color = 33
	} else {
		color = 32
	}

	log.Printf("|\033[7;%dm %d \033[0m| %s\n", color, status, msg)
}

func writeCORS(rw http.ResponseWriter) {
	if len(conf.AllowOrigin) > 0 {
		rw.Header().Set("Access-Control-Allow-Origin", conf.AllowOrigin)
		rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONs")
	}
}

func respondWithImage(reqID string, r *http.Request, rw http.ResponseWriter, data []byte, imgURL string, po processingOptions, duration time.Duration) {
	gzipped := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && conf.GZipCompression > 0

	rw.Header().Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))
	rw.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public", conf.TTL))
	rw.Header().Set("Content-Type", mimes[po.Format])

	dataToRespond := data

	if gzipped {
		var buf bytes.Buffer

		gz, _ := gzip.NewWriterLevel(&buf, conf.GZipCompression)
		gz.Write(data)
		gz.Close()

		dataToRespond = buf.Bytes()

		rw.Header().Set("Content-Encoding", "gzip")
	}

	rw.Header().Set("Content-Length", strconv.Itoa(len(dataToRespond)))

	rw.WriteHeader(200)
	rw.Write(dataToRespond)

	logResponse(200, fmt.Sprintf("[%s] Processed in %s: %s; %+v", reqID, duration, imgURL, po))
}

func respondWithError(reqID string, rw http.ResponseWriter, err imgproxyError) {
	logResponse(err.StatusCode, fmt.Sprintf("[%s] %s", reqID, err.Message))

	rw.WriteHeader(err.StatusCode)
	rw.Write([]byte(err.PublicMessage))
}

func respondWithOptions(reqID string, rw http.ResponseWriter) {
	logResponse(200, fmt.Sprintf("[%s] Respond with options", reqID))
	rw.WriteHeader(200)
}

func checkSecret(s string) bool {
	if len(conf.Secret) == 0 {
		return true
	}
	return strings.HasPrefix(s, "Bearer ") && subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(s, "Bearer ")), []byte(conf.Secret)) == 1
}

func (h *httpHandler) lock() {
	h.sem <- struct{}{}
}

func (h *httpHandler) unlock() {
	<-h.sem
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	reqID, _ := nanoid.Nanoid()

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(imgproxyError); ok {
				respondWithError(reqID, rw, err)
			} else {
				respondWithError(reqID, rw, newUnexpectedError(r.(error), 4))
			}
		}
	}()

	log.Printf("[%s] %s: %s\n", reqID, r.Method, r.URL.RequestURI())

	writeCORS(rw)

	if r.Method == http.MethodOptions {
		respondWithOptions(reqID, rw)
		return
	}

	if r.Method != http.MethodGet {
		panic(invalidMethodErr)
	}

	if !checkSecret(r.Header.Get("Authorization")) {
		panic(invalidSecretErr)
	}

	h.lock()
	defer h.unlock()

	if r.URL.Path == "/health" {
		rw.WriteHeader(200)
		rw.Write([]byte("imgproxy is running"))
		return
	}

	t := startTimer(time.Duration(conf.WriteTimeout)*time.Second, "Processing")

	imgURL, procOpt, err := parsePath(r)
	if err != nil {
		panic(newError(404, err.Error(), "Invalid image url"))
	}

	if _, err = url.ParseRequestURI(imgURL); err != nil {
		panic(newError(404, err.Error(), "Invalid image url"))
	}

	b, imgtype, err := downloadImage(imgURL)
	if err != nil {
		panic(newError(404, err.Error(), "Image is unreachable"))
	}

	t.Check()

	if conf.ETagEnabled {
		eTag := calcETag(b, &procOpt)
		rw.Header().Set("ETag", eTag)

		if eTag == r.Header.Get("If-None-Match") {
			panic(notModifiedErr)
		}
	}

	t.Check()

	b, err = processImage(b, imgtype, procOpt, t)
	if err != nil {
		panic(newError(500, err.Error(), "Error occurred while processing image"))
	}

	t.Check()

	respondWithImage(reqID, r, rw, b, imgURL, procOpt, t.Since())
}
