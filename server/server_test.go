package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite
	config      *Config
	blankRouter *Router
}

func (s *ServerTestSuite) SetupTest() {
	config.Reset()
	s.config = NewConfigFromEnv()
	s.config.Bind = "127.0.0.1:0" // Use port 0 for auto-assignment
	s.blankRouter = NewRouter(s.config)
}

func (s *ServerTestSuite) mockHandler(reqID string, rw http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *ServerTestSuite) TestStartServerWithInvalidBind() {
	ctx, cancel := context.WithCancel(s.T().Context())

	// Track if cancel was called using atomic
	var cancelCalled atomic.Bool
	cancelWrapper := func() {
		cancel()
		cancelCalled.Store(true)
	}

	invalidConfig := &Config{
		Network: "tcp",
		Bind:    "invalid-address", // Invalid address
	}

	r := NewRouter(invalidConfig)

	server, err := Start(cancelWrapper, r)

	s.Require().Error(err)
	s.Nil(server)
	s.Contains(err.Error(), "can't start server")

	// Check if cancel was called using Eventually
	s.Require().Eventually(cancelCalled.Load, 100*time.Millisecond, 10*time.Millisecond)

	// Also verify the context was cancelled
	s.Require().Eventually(func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}, 100*time.Millisecond, 10*time.Millisecond)
}

func (s *ServerTestSuite) TestShutdown() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := Start(cancel, s.blankRouter)
	s.Require().NoError(err)
	s.NotNil(server)

	// Test graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(s.T().Context(), 10*time.Second)
	defer shutdownCancel()

	// Should not panic or hang
	s.NotPanics(func() {
		server.Shutdown(shutdownCtx)
	})
}

func (s *ServerTestSuite) TestWithCORS() {
	tests := []struct {
		name            string
		corsAllowOrigin string
		expectedOrigin  string
		expectedMethods string
	}{
		{
			name:            "WithCORSOrigin",
			corsAllowOrigin: "https://example.com",
			expectedOrigin:  "https://example.com",
			expectedMethods: "GET, OPTIONS",
		},
		{
			name:            "NoCORSOrigin",
			corsAllowOrigin: "",
			expectedOrigin:  "",
			expectedMethods: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			config := &Config{
				CORSAllowOrigin: tt.corsAllowOrigin,
			}
			router := NewRouter(config)

			wrappedHandler := router.WithCORS(s.mockHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rw := httptest.NewRecorder()

			wrappedHandler("test-req-id", rw, req)

			s.Equal(tt.expectedOrigin, rw.Header().Get(httpheaders.AccessControlAllowOrigin))
			s.Equal(tt.expectedMethods, rw.Header().Get(httpheaders.AccessControlAllowMethods))
		})
	}
}

func (s *ServerTestSuite) TestWithSecret() {
	tests := []struct {
		name        string
		secret      string
		authHeader  string
		expectError bool
	}{
		{
			name:       "ValidSecret",
			secret:     "test-secret",
			authHeader: "Bearer test-secret",
		},
		{
			name:        "InvalidSecret",
			secret:      "foo-secret",
			authHeader:  "Bearer wrong-secret",
			expectError: true,
		},
		{
			name:       "NoSecretConfigured",
			secret:     "",
			authHeader: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			config := &Config{
				Secret: tt.secret,
			}
			router := NewRouter(config)

			wrappedHandler := router.WithSecret(s.mockHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set(httpheaders.Authorization, tt.authHeader)
			}
			rw := httptest.NewRecorder()

			err := wrappedHandler("test-req-id", rw, req)

			if tt.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ServerTestSuite) TestIntoSuccess() {
	mockHandler := func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		rw.WriteHeader(http.StatusOK)
		return nil
	}

	wrappedHandler := s.blankRouter.WithReportError(mockHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	wrappedHandler("test-req-id", rw, req)

	s.Equal(http.StatusOK, rw.Code)
}

func (s *ServerTestSuite) TestIntoWithError() {
	testError := errors.New("test error")
	mockHandler := func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		return testError
	}

	wrappedHandler := s.blankRouter.WithReportError(mockHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	wrappedHandler("test-req-id", rw, req)

	s.Equal(http.StatusInternalServerError, rw.Code)
	s.Equal("text/plain", rw.Header().Get(httpheaders.ContentType))
}

func (s *ServerTestSuite) TestIntoPanicWithError() {
	testError := errors.New("panic error")
	mockHandler := func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		panic(testError)
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	s.NotPanics(func() {
		err := wrappedHandler("test-req-id", rw, req)
		s.Require().Error(err, "panic error")
	})

	s.Equal(http.StatusOK, rw.Code)
}

func (s *ServerTestSuite) TestIntoPanicWithAbortHandler() {
	mockHandler := func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		panic(http.ErrAbortHandler)
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	// Should re-panic with ErrAbortHandler
	s.Panics(func() {
		wrappedHandler("test-req-id", rw, req)
	})
}

func (s *ServerTestSuite) TestIntoPanicWithNonError() {
	mockHandler := func(reqID string, rw http.ResponseWriter, r *http.Request) error {
		panic("string panic")
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	// Should re-panic with non-error panics
	s.NotPanics(func() {
		err := wrappedHandler("test-req-id", rw, req)
		s.Require().Error(err, "string panic")
	})
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
