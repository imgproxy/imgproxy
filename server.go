package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"
	"strings"
	"golang.org/x/net/netutil"
)

var (
	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidSecret = newError(403, "Invalid secret", "Forbidden")
)

func buildRouter() *router {
	r := newRouter(conf.PathPrefix)

	r.PanicHandler = handlePanic

	r.GET("/", handleLanding, true)
	r.GET("/health", handleHealth, true)
	r.GET("/favicon.ico", handleFavicon, true)
	r.GET("/", withCORS(withSecret(handleProcessing)), false)
	r.HEAD("/", withCORS(handleHead), false)
	r.OPTIONS("/", withCORS(handleHead), false)

	return r
}

func startServer(cancel context.CancelFunc) (*http.Server, error) {
	l, err := listenReuseport(conf.Network, conf.Bind)
	if err != nil {
		return nil, fmt.Errorf("Can't start server: %s", err)
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

	if err := initProcessingHandler(); err != nil {
		return nil, err
	}

	go func() {
		logNotice("Starting server at %s", conf.Bind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			logError(err.Error())
		}
		cancel()
	}()

	return s, nil
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
			origins := strings.Split(conf.AllowOrigin, " ")
			requestOrigin := r.Header.Get("Origin")

			for _,origin := range origins {
				if(strings.TrimSpace(origin) == requestOrigin) {
					rw.Header().Set("Access-Control-Allow-Origin", requestOrigin)
					rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
					rw.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}
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
	var (
		ierr *imgproxyError
		ok   bool
	)

	if ierr, ok = err.(*imgproxyError); !ok {
		ierr = newUnexpectedError(err.Error(), 3)
	}

	if ierr.Unexpected {
		reportError(err, r)
	}

	logResponse(reqID, r, ierr.StatusCode, ierr, nil, nil)

	rw.WriteHeader(ierr.StatusCode)

	if conf.DevelopmentErrorsMode {
		rw.Write([]byte(ierr.Message))
	} else {
		rw.Write([]byte(ierr.PublicMessage))
	}
}

func handleHealth(reqID string, rw http.ResponseWriter, r *http.Request) {
	logResponse(reqID, r, 200, nil, nil, nil)
	rw.WriteHeader(200)
	rw.Write(imgproxyIsRunningMsg)
}

func handleHead(reqID string, rw http.ResponseWriter, r *http.Request) {
	logResponse(reqID, r, 200, nil, nil, nil)
	rw.WriteHeader(200)
}

func handleFavicon(reqID string, rw http.ResponseWriter, r *http.Request) {
	logResponse(reqID, r, 200, nil, nil, nil)
	// TODO: Add a real favicon maybe?
	rw.WriteHeader(200)
}
