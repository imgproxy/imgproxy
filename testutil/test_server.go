package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/stretchr/testify/require"
)

// TestServerHookFunc is a function type for in-request hooks
type TestServerHookFunc func(r *http.Request, rw http.ResponseWriter)

// Sugar alias
type LazyTestServer = LazyObj[*TestServer]

// TestServer is a syntax sugar wrapper over httptest.Server
type TestServer struct {
	testServer *httptest.Server
	status     int
	data       []byte
	header     http.Header
	hook       TestServerHookFunc
}

// NewLazySuiteTestServer creates a lazy TestServer object for use in test suites
func NewLazySuiteTestServer(
	l LazySuiteFrom,
	init ...func(*TestServer) error,
) (LazyObj[*TestServer], context.CancelFunc) {
	return NewLazySuiteObj(
		l,
		func() (*TestServer, error) {
			s := NewTestServer()

			if len(init) > 0 {
				for _, fn := range init {
					if fn == nil {
						continue
					}

					err := fn(s)
					require.NoError(l.Lazy().T(), err, "Failed to reset test server")
				}
			}

			return s, nil
		},
		func(s *TestServer) error {
			s.Close()
			return nil
		},
	)
}

// New creates and starts new http.TestServer
func NewTestServer() *TestServer {
	ts := &TestServer{
		status: http.StatusOK,
		header: make(http.Header),
		data:   nil,
		hook:   nil,
	}

	return ts.start()
}

// SetStatusCode sets the status code that will be returned by the server
func (s *TestServer) SetStatusCode(status int) *TestServer {
	s.status = status
	return s
}

// SetBody sets the body that will be returned by the server
func (s *TestServer) SetBody(data []byte) *TestServer {
	s.data = data
	return s
}

// WithHeader adds headers that will be returned by the server.
// Odd arguments are treated as keys, even arguments as values.
func (s *TestServer) SetHeaders(kv ...string) *TestServer {
	for i := 0; i+1 < len(kv); i += 2 {
		key := kv[i]
		value := kv[i+1]
		s.header.Set(key, value)
	}

	return s
}

// SetHook sets a function that will be called on each request. It is called
// after headsers are set, but before status and body are written.
func (s *TestServer) SetHook(f TestServerHookFunc) *TestServer {
	s.hook = f
	return s
}

// Start starts the server
func (s *TestServer) start() *TestServer {
	s.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpheaders.CopyAll(s.header, w.Header(), true)
		if s.hook != nil {
			s.hook(r, w)
		}
		w.WriteHeader(s.status)
		w.Write(s.data)
	}))

	return s
}

// Close stops the server
func (s *TestServer) Close() {
	s.testServer.Close()
}

// URL returns the server URL
func (s *TestServer) URL() string {
	return s.testServer.URL
}
