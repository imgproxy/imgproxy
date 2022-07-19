package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	golog "log"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/reuseport"
	"github.com/imgproxy/imgproxy/v3/router"
)

var (
	imgproxyIsRunningMsg = []byte("imgproxy is running")

	errInvalidSecret = ierrors.New(403, "Invalid secret", "Forbidden")
)

func buildRouter() *router.Router {
	r := router.New(config.PathPrefix)

	r.GET("/", handleLanding, true)
	r.GET("/health", handleHealth, true)
	if len(config.HealthCheckPath) > 0 {
		r.GET(config.HealthCheckPath, handleHealth, true)
	}
	r.GET("/favicon.ico", handleFavicon, true)
	r.GET("/", withMetrics(withPanicHandler(withCORS(withSecret(handleProcessing)))), false)
	r.HEAD("/", withCORS(handleHead), false)
	r.OPTIONS("/", withCORS(handleHead), false)

	return r
}

func startServer(cancel context.CancelFunc) (*http.Server, error) {
	l, err := reuseport.Listen(config.Network, config.Bind)
	if err != nil {
		return nil, fmt.Errorf("Can't start server: %s", err)
	}

	if config.MaxClients > 0 {
		l = netutil.LimitListener(l, config.MaxClients)
	}

	errLogger := golog.New(
		log.WithField("source", "http_server").WriterLevel(log.ErrorLevel),
		"", 0,
	)

	s := &http.Server{
		Handler:        buildRouter(),
		ReadTimeout:    time.Duration(config.ReadTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       errLogger,
	}

	if config.KeepAliveTimeout > 0 {
		s.IdleTimeout = time.Duration(config.KeepAliveTimeout) * time.Second
	} else {
		s.SetKeepAlivesEnabled(false)
	}

	go func() {
		log.Infof("Starting server at %s", config.Bind)
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}
		cancel()
	}()

	return s, nil
}

func shutdownServer(s *http.Server) {
	log.Info("Shutting down the server...")

	ctx, close := context.WithTimeout(context.Background(), 5*time.Second)
	defer close()

	s.Shutdown(ctx)
}

func withMetrics(h router.RouteHandler) router.RouteHandler {
	if !metrics.Enabled() {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		ctx, metricsCancel, rw := metrics.StartRequest(r.Context(), rw, r)
		defer metricsCancel()

		h(reqID, rw, r.WithContext(ctx))
	}
}

func withCORS(h router.RouteHandler) router.RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		if len(config.AllowOrigin) > 0 {
			rw.Header().Set("Access-Control-Allow-Origin", config.AllowOrigin)
			rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}

		h(reqID, rw, r)
	}
}

func withSecret(h router.RouteHandler) router.RouteHandler {
	if len(config.Secret) == 0 {
		return h
	}

	authHeader := []byte(fmt.Sprintf("Bearer %s", config.Secret))

	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), authHeader) == 1 {
			h(reqID, rw, r)
		} else {
			panic(errInvalidSecret)
		}
	}
}

func withPanicHandler(h router.RouteHandler) router.RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if rerr := recover(); rerr != nil {
				err, ok := rerr.(error)
				if !ok {
					panic(rerr)
				}

				ierr := ierrors.Wrap(err, 2)

				if ierr.Unexpected {
					errorreport.Report(err, r)
				}

				router.LogResponse(reqID, r, ierr.StatusCode, ierr)

				rw.WriteHeader(ierr.StatusCode)

				if config.DevelopmentErrorsMode {
					rw.Write([]byte(ierr.Message))
				} else {
					rw.Write([]byte(ierr.PublicMessage))
				}
			}
		}()

		h(reqID, rw, r)
	}
}

func handleHealth(reqID string, rw http.ResponseWriter, r *http.Request) {
	router.LogResponse(reqID, r, 200, nil)
	rw.WriteHeader(200)
	rw.Write(imgproxyIsRunningMsg)
}

func handleHead(reqID string, rw http.ResponseWriter, r *http.Request) {
	router.LogResponse(reqID, r, 200, nil)
	rw.WriteHeader(200)
}

func handleFavicon(reqID string, rw http.ResponseWriter, r *http.Request) {
	router.LogResponse(reqID, r, 200, nil)
	// TODO: Add a real favicon maybe?
	rw.WriteHeader(200)
}
