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
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/reuseport"
	"github.com/imgproxy/imgproxy/v3/server"
)

const (
	faviconPath = "/favicon.ico"
	healthPath  = "/health"
)

func buildRouter() *server.Router {
	r := server.NewRouter(config.PathPrefix)

	r.GET("/", true, handlers.LandingHandler)
	r.GET("", true, handlers.LandingHandler)

	r.GET("/", false, handleProcessing, withMetrics, withPanicHandler, withCORS, withSecret)

	r.HEAD("/", false, r.OkHandler, withCORS)
	r.OPTIONS("/", false, r.OkHandler, withCORS)

	r.GET(faviconPath, true, r.NotFoundHandler).Silent()
	r.GET(healthPath, true, handlers.HealthHandler).Silent()
	if config.HealthCheckPath != "" {
		r.GET(config.HealthCheckPath, true, handlers.HealthHandler).Silent()
	}

	return r
}

func startServer(cancel context.CancelFunc) (*http.Server, error) {
	l, err := reuseport.Listen(config.Network, config.Bind)
	if err != nil {
		return nil, fmt.Errorf("can't start server: %s", err)
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
		ReadTimeout:    time.Duration(config.ReadRequestTimeout) * time.Second,
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

func withMetrics(h server.RouteHandler) server.RouteHandler {
	if !metrics.Enabled() {
		return h
	}

	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		ctx, metricsCancel, rw := metrics.StartRequest(r.Context(), rw, r)
		defer metricsCancel()

		h(reqID, rw, r.WithContext(ctx))

		return nil
	}
}

func withCORS(h server.RouteHandler) server.RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		if len(config.AllowOrigin) > 0 {
			rw.Header().Set("Access-Control-Allow-Origin", config.AllowOrigin)
			rw.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		}

		h(reqID, rw, r)

		return nil
	}
}

func withSecret(h server.RouteHandler) server.RouteHandler {
	if len(config.Secret) == 0 {
		return h
	}

	authHeader := []byte(fmt.Sprintf("Bearer %s", config.Secret))

	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), authHeader) == 1 {
			h(reqID, rw, r)
		} else {
			panic(newInvalidSecretError())
		}

		return nil
	}
}

func withPanicHandler(h server.RouteHandler) server.RouteHandler {
	return func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		ctx := errorreport.StartRequest(r)
		r = r.WithContext(ctx)

		errorreport.SetMetadata(r, "Request ID", reqID)

		defer func() {
			if rerr := recover(); rerr != nil {
				if rerr == http.ErrAbortHandler {
					panic(rerr)
				}

				err, ok := rerr.(error)
				if !ok {
					panic(rerr)
				}

				ierr := ierrors.Wrap(err, 0)

				if ierr.ShouldReport() {
					errorreport.Report(err, r)
				}

				server.LogResponse(reqID, r, ierr.StatusCode(), ierr)

				rw.Header().Set("Content-Type", "text/plain")
				rw.WriteHeader(ierr.StatusCode())

				if config.DevelopmentErrorsMode {
					rw.Write([]byte(ierr.Error()))
				} else {
					rw.Write([]byte(ierr.PublicMessage()))
				}
			}
		}()

		h(reqID, rw, r)

		return nil
	}
}
