package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	nanoid "github.com/matoous/go-nanoid"
	"github.com/valyala/fasthttp"
)

var (
	mimes = map[imageType]string{
		imageTypeJPEG: "image/jpeg",
		imageTypePNG:  "image/png",
		imageTypeWEBP: "image/webp",
	}

	responseBufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	authHeaderMust []byte

	healthRequestURI = []byte("/health")

	serverMutex mutex
)

func startServer() *fasthttp.Server {
	serverMutex = newMutex(conf.Concurrency)

	s := &fasthttp.Server{
		Name:        "imgproxy",
		Handler:     serveHTTP,
		Concurrency: conf.MaxClients,
		ReadTimeout: time.Duration(conf.ReadTimeout) * time.Second,
	}

	go func() {
		log.Printf("Starting server at %s\n", conf.Bind)
		if err := s.ListenAndServe(conf.Bind); err != nil {
			log.Fatalln(err)
		}
	}()

	return s
}

func shutdownServer(s *fasthttp.Server) {
	log.Println("Shutting down the server...")
	s.Shutdown()
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

func writeCORS(rctx *fasthttp.RequestCtx) {
	if len(conf.AllowOrigin) > 0 {
		rctx.Request.Header.Set("Access-Control-Allow-Origin", conf.AllowOrigin)
		rctx.Request.Header.Set("Access-Control-Allow-Methods", "GET, OPTIONs")
	}
}

func respondWithImage(ctx context.Context, reqID string, rctx *fasthttp.RequestCtx, data []byte) {
	rctx.SetStatusCode(200)

	po := getprocessingOptions(ctx)

	rctx.SetContentType(mimes[po.Format])
	rctx.Response.Header.Set("Cache-Control", fmt.Sprintf("max-age=%d, public", conf.TTL))
	rctx.Response.Header.Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))

	if conf.GZipCompression > 0 && rctx.Request.Header.HasAcceptEncodingBytes([]byte("gzip")) {
		rctx.Response.Header.Set("Content-Encoding", "gzip")
		gzipData(data, rctx)
	} else {
		rctx.SetBody(data)
	}

	logResponse(200, fmt.Sprintf("[%s] Processed in %s: %s; %+v", reqID, getTimerSince(ctx), getImageURL(ctx), po))
}

func respondWithError(reqID string, rctx *fasthttp.RequestCtx, err imgproxyError) {
	logResponse(err.StatusCode, fmt.Sprintf("[%s] %s", reqID, err.Message))

	rctx.SetStatusCode(err.StatusCode)
	rctx.SetBodyString(err.PublicMessage)
}

func respondWithOptions(reqID string, rctx *fasthttp.RequestCtx) {
	logResponse(200, fmt.Sprintf("[%s] Respond with options", reqID))
	rctx.SetStatusCode(200)
}

func prepareAuthHeaderMust() []byte {
	if len(authHeaderMust) == 0 {
		buf := bytes.NewBufferString("Bearer ")
		buf.WriteString(conf.Secret)
		authHeaderMust = []byte(conf.Secret)
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

func serveHTTP(rctx *fasthttp.RequestCtx) {
	reqID, _ := nanoid.Nanoid()

	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(imgproxyError); ok {
				respondWithError(reqID, rctx, err)
			} else {
				respondWithError(reqID, rctx, newUnexpectedError(r.(error), 4))
			}
		}
	}()

	log.Printf("[%s] %s: %s\n", reqID, rctx.Method(), rctx.RequestURI())

	writeCORS(rctx)

	if rctx.Request.Header.IsOptions() {
		respondWithOptions(reqID, rctx)
		return
	}

	if !rctx.IsGet() {
		panic(invalidMethodErr)
	}

	if !checkSecret(rctx) {
		panic(invalidSecretErr)
	}

	serverMutex.Lock()
	defer serverMutex.Unock()

	if bytes.Equal(rctx.RequestURI(), healthRequestURI) {
		rctx.SetStatusCode(200)
		rctx.SetBodyString("imgproxy is running")
		return
	}

	ctx, timeoutCancel := startTimer(time.Duration(conf.WriteTimeout) * time.Second)
	defer timeoutCancel()

	ctx, err := parsePath(ctx, rctx)
	if err != nil {
		panic(newError(404, err.Error(), "Invalid image url"))
	}

	ctx, downloadcancel, err := downloadImage(ctx)
	defer downloadcancel()
	if err != nil {
		panic(newError(404, err.Error(), "Image is unreachable"))
	}

	checkTimeout(ctx)

	// if conf.ETagEnabled {
	// 	eTag := calcETag(b, &procOpt)
	// 	rw.Header().Set("ETag", eTag)

	// 	if eTag == r.Header.Get("If-None-Match") {
	// 		panic(notModifiedErr)
	// 	}
	// }

	checkTimeout(ctx)

	imageData, err := processImage(ctx)
	if err != nil {
		panic(newError(500, err.Error(), "Error occurred while processing image"))
	}

	checkTimeout(ctx)

	respondWithImage(ctx, reqID, rctx, imageData)
}
