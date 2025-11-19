package s3

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

// TestServer is a mock S3 server for testing
type TestServer struct {
	server  *httptest.Server
	backend *s3mem.Backend
}

// Backend returns the underlying s3mem.Backend for direct API access
func (s *TestServer) Backend() *s3mem.Backend {
	return s.backend
}

// s3StorageWrapper wraps the storage and optionally holds a server for cleanup
type s3StorageWrapper struct {
	*Storage
	server      *TestServer
	shouldClose bool
}

// Server returns the underlying S3Server
func (w *s3StorageWrapper) Server() *TestServer {
	return w.server
}

// Sugar alias
type LazySuiteStorage = testutil.LazyObj[*s3StorageWrapper]

// NewLazySuiteStorage creates a lazy S3 Storage object for use in test suites
// A new server will be created internally and cleaned up automatically
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
) (testutil.LazyObj[*s3StorageWrapper], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*s3StorageWrapper, error) {
			wrapper := &s3StorageWrapper{}

			// Create server internally
			s3Server := NewS3Server()
			wrapper.server = s3Server
			wrapper.shouldClose = true

			// Create bucket first using backend directly
			err := s3Server.backend.CreateBucket("test-container")
			if err != nil {
				return nil, err
			}

			os.Setenv("AWS_ACCESS_KEY_ID", "TEST")
			os.Setenv("AWS_SECRET_ACCESS_KEY", "TEST")
			os.Setenv("AWS_REGION", "us-east-1")

			config := NewDefaultConfig()
			config.Endpoint = s3Server.URL()
			config.Region = "us-east-1"
			config.EndpointUsePathStyle = true

			storage, err := New(&config, http.DefaultTransport.(*http.Transport))
			if err != nil {
				return nil, err
			}

			wrapper.Storage = storage
			return wrapper, nil
		},
		func(w *s3StorageWrapper) error {
			// Clean up internal server if we created it
			if w.shouldClose {
				w.server.Close()
			}
			return nil
		},
	)
}

// NewS3Server creates and starts a new mock S3 server
func NewS3Server() *TestServer {
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	server := httptest.NewServer(faker.Server())

	return &TestServer{
		server:  server,
		backend: backend,
	}
}

// Close stops the server
func (s *TestServer) Close() {
	s.server.Close()
}

// URL returns the server URL
func (s *TestServer) URL() string {
	return s.server.URL
}
