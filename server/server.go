package server

import (
	"context"
	"fmt"
	golog "log"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/netutil"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/reuseport"
)

const (
	// maxHeaderBytes represents max bytes in request header
	maxHeaderBytes = 1 << 20
)

// Server represents the HTTP server wrapper struct
type Server struct {
	router *Router
	server *http.Server
	Addr   net.Addr
}

// Start starts the http server. cancel is called in case server failed to start, but it happened
// asynchronously. It should cancel the upstream context.
func Start(cancel context.CancelFunc, router *Router) (*Server, error) {
	l, err := reuseport.Listen(router.config.Network, router.config.Bind, router.config.SocketReusePort)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("can't start server: %s", err)
	}

	if router.config.MaxClients > 0 {
		l = netutil.LimitListener(l, router.config.MaxClients)
	}

	errLogger := golog.New(
		log.WithField("source", "http_server").WriterLevel(log.ErrorLevel),
		"", 0,
	)

	addr := l.Addr()

	srv := &http.Server{
		Handler:        router,
		ReadTimeout:    router.config.ReadRequestTimeout,
		MaxHeaderBytes: maxHeaderBytes,
		ErrorLog:       errLogger,
	}

	if config.KeepAliveTimeout > 0 {
		srv.IdleTimeout = router.config.KeepAliveTimeout
	} else {
		srv.SetKeepAlivesEnabled(false)
	}

	go func() {
		log.Infof("Starting server at %s", router.config.Bind)

		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Error(err)
		}

		cancel()
	}()

	return &Server{
		router: router,
		server: srv,
		Addr:   addr,
	}, nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) {
	log.Info("Shutting down the server...")

	ctx, close := context.WithTimeout(ctx, s.router.config.GracefulTimeout)
	defer close()

	s.server.Shutdown(ctx)
}
