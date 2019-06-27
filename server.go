package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/netutil"
)

var (
	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidSecret = newError(403, "Invalid secret", "Forbidden")
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
	l, err := listenReuseport("tcp", conf.Bind)
	if err != nil {
		logFatal(err.Error())
	}
	l = netutil.LimitListener(l, conf.MaxClients)

	s := &http.Server{
		Handler:        buildRouter(),
		ReadTimeout:    time.Duration(conf.ReadTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if conf.KeepAliveTimeout > 0 {
		s.IdleTimeout = time.Duration(conf.KeepAliveTimeout) * time.Second
	} else {
		s.SetKeepAlivesEnabled(false)
	}

	initProcessingHandler()

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
			panic(errInvalidSecret)
		}
	}
}

func handlePanic(reqID string, rw http.ResponseWriter, r *http.Request, err error) {
	reportError(err, r)

	var (
		ierr *imgproxyError
		ok   bool
	)

	if ierr, ok = err.(*imgproxyError); !ok {
		ierr = newUnexpectedError(err.Error(), 3)
	}

	logResponse(reqID, ierr.StatusCode, ierr.Message)

	rw.WriteHeader(ierr.StatusCode)

	if conf.DevelopmentErrorsMode {
		rw.Write([]byte(ierr.Message))
	} else {
		rw.Write([]byte(ierr.PublicMessage))
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
