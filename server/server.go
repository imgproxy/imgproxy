package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

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

	errLogger := slog.NewLogLogger(
		slog.With("source", "http_server").Handler(),
		slog.LevelError,
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
		slog.Info(fmt.Sprintf("Starting server at %s", router.config.Bind))

		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error(), "source", "http_server")
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
	slog.Info("Shutting down the server...")

	ctx, close := context.WithTimeout(ctx, s.router.config.GracefulTimeout)
	defer close()

	s.server.Shutdown(ctx)
}
