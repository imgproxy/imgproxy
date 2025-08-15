package server

import (
	"context"
	"fmt"
	golog "log"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"

	"github.com/imgproxy/imgproxy/v3/reuseport"
)

const (
	// maxHeaderBytes represents max bytes in request header
	maxHeaderBytes = 1 << 20
)

// Server represents the HTTP server wrapper struct
type Server struct {
	router *Router
	config *Config
	server *http.Server
}

// Start starts the http server. cancel is called in case server failed to start, but it happened
// asynchronously. It should cancel the upstream context.
func Start(cancel context.CancelFunc, router *Router, config *Config) (*Server, error) {
	l, err := reuseport.Listen(config.Network, config.Bind, config.SocketReusePort)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can't start server: %s", err)
	}

	if config.MaxClients > 0 {
		l = netutil.LimitListener(l, config.MaxClients)
	}

	errLogger := golog.New(
		log.WithField("source", "http_server").WriterLevel(log.ErrorLevel),
		"", 0,
	)

	srv := &http.Server{
		Handler:        router,
		ReadTimeout:    config.ReadRequestTimeout,
		MaxHeaderBytes: maxHeaderBytes,
		ErrorLog:       errLogger,
	}

	if config.KeepAliveTimeout > 0 {
		srv.IdleTimeout = config.KeepAliveTimeout
	} else {
		srv.SetKeepAlivesEnabled(false)
	}

	go func() {
		log.Infof("Starting server at %s", config.Bind)

		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}

		cancel()
	}()

	return &Server{
		router: router,
		config: config,
		server: srv,
	}, nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) {
	log.Info("Shutting down the server...")

	ctx, close := context.WithTimeout(ctx, s.config.GracefulTimeout)
	defer close()

	s.server.Shutdown(ctx)
}
