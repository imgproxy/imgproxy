package swift

import (
	"context"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
)

// TestServer is a mock Swift server for testing
type TestServer struct {
	server     *swifttest.SwiftServer
	connection *swift.Connection
}

// swiftStorageWrapper wraps the storage and optionally holds a server for cleanup
type swiftStorageWrapper struct {
	*Storage
	server      *TestServer
	connection  *swift.Connection
	shouldClose bool
}

// Server returns the underlying SwiftServer
func (w *swiftStorageWrapper) Server() *TestServer {
	return w.server
}

// Connection returns the Swift connection for direct API access
func (w *swiftStorageWrapper) Connection() *swift.Connection {
	return w.connection
}

// Sugar alias
type LazySuiteStorage = testutil.LazyObj[*swiftStorageWrapper]

// NewLazySuiteStorage creates a lazy Swift Storage object for use in test suites
// A new server will be created internally and cleaned up automatically
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
) (testutil.LazyObj[*swiftStorageWrapper], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*swiftStorageWrapper, error) {
			wrapper := &swiftStorageWrapper{}

			// Create server internally
			swiftServer, err := NewSwiftServer()
			if err != nil {
				return nil, err
			}
			wrapper.server = swiftServer
			wrapper.shouldClose = true

			// Create container first using connection directly
			err = swiftServer.connection.ContainerCreate(context.Background(), "test-container", nil)
			if err != nil {
				return nil, err
			}

			config := NewDefaultConfig()
			config.AuthURL = swiftServer.server.AuthURL
			config.Username = swifttest.TEST_ACCOUNT
			config.APIKey = swifttest.TEST_ACCOUNT
			config.AuthVersion = 1

			storage, err := New(l.Lazy().T().Context(), &config, http.DefaultTransport.(*http.Transport))
			if err != nil {
				return nil, err
			}

			wrapper.Storage = storage
			wrapper.connection = swiftServer.connection
			return wrapper, nil
		},
		func(w *swiftStorageWrapper) error {
			// Clean up internal server if we created it
			if w.shouldClose {
				w.server.Close()
			}
			return nil
		},
	)
}

// NewSwiftServer creates and starts a new mock Swift server
func NewSwiftServer() (*TestServer, error) {
	server, err := swifttest.NewSwiftServer("localhost")
	if err != nil {
		return nil, err
	}

	// Create connection
	conn := &swift.Connection{
		UserName:    swifttest.TEST_ACCOUNT,
		ApiKey:      swifttest.TEST_ACCOUNT,
		AuthUrl:     server.AuthURL,
		AuthVersion: 1,
	}

	// Authenticate
	err = conn.Authenticate(context.Background())
	if err != nil {
		server.Close()
		return nil, err
	}

	return &TestServer{
		server:     server,
		connection: conn,
	}, nil
}

// Connection returns the Swift connection
func (s *TestServer) Connection() *swift.Connection {
	return s.connection
}

// Close stops the server
func (s *TestServer) Close() {
	s.server.Close()
}

// URL returns the server auth URL
func (s *TestServer) URL() string {
	return s.server.AuthURL
}
