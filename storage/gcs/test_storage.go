package gcs

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

// TestServer is a mock Google Cloud Storage server for testing
type TestServer struct {
	server *fakestorage.Server
}

// gcsStorageWrapper wraps the storage and optionally holds a server for cleanup
type gcsStorageWrapper struct {
	*Storage
	server      *TestServer
	shouldClose bool
}

// Server returns the underlying GcsServer
func (w *gcsStorageWrapper) Server() *TestServer {
	return w.server
}

// Sugar alias
type LazySuiteStorage = testutil.LazyObj[*gcsStorageWrapper]

// NewTestServer creates and starts a new mock GCS server
func NewTestServer() (*TestServer, error) {
	port, err := getFreePort()
	if err != nil {
		return nil, err
	}

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		Scheme:     "http",
		Port:       uint16(port),
		PublicHost: fmt.Sprintf("localhost:%d", port),
	})
	if err != nil {
		return nil, err
	}

	return &TestServer{
		server: server,
	}, nil
}

// getFreePort finds an available TCP port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Close stops the server
func (s *TestServer) Close() {
	s.server.Stop()
}

// URL returns the server URL for storage API
func (s *TestServer) URL() string {
	return s.server.URL() + "/storage/v1/"
}

// PublicURL returns the public server URL (for storage_test.go compatibility)
func (s *TestServer) PublicURL() string {
	return s.server.PublicURL()
}

// Server returns the underlying fake storage server
func (s *TestServer) Server() *fakestorage.Server {
	return s.server
}

// NewLazySuiteStorage creates a lazy GCS Storage object for use in test suites
// A new server will be created internally with optional initial objects and cleaned up automatically
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
	initialObjects []fakestorage.Object,
) (testutil.LazyObj[*gcsStorageWrapper], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*gcsStorageWrapper, error) {
			wrapper := &gcsStorageWrapper{}

			// Create server internally with optional initial objects
			port, err := getFreePort()
			if err != nil {
				return nil, err
			}

			server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
				Scheme:         "http",
				Port:           uint16(port),
				PublicHost:     fmt.Sprintf("localhost:%d", port),
				InitialObjects: initialObjects,
			})
			if err != nil {
				return nil, err
			}

			gcsServer := &TestServer{
				server: server,
			}
			wrapper.server = gcsServer
			wrapper.shouldClose = true

			config := NewDefaultConfig()
			config.Endpoint = gcsServer.PublicURL() + "/storage/v1/"
			config.TestNoAuth = true

			storage, err := New(&config, http.DefaultTransport.(*http.Transport))
			if err != nil {
				return nil, err
			}

			wrapper.Storage = storage
			return wrapper, nil
		},
		func(w *gcsStorageWrapper) error {
			// Clean up internal server if we created it
			if w.shouldClose {
				w.server.Close()
			}
			return nil
		},
	)
}
