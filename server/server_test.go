package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/stretchr/testify/suite"
)

type ServerTestSuite struct {
	suite.Suite

	config        *Config
	monitoring    *monitoring.Monitoring
	errorReporter *errorreport.Reporter
	blankRouter   *Router
}

func (s *ServerTestSuite) SetupTest() {
	c := NewDefaultConfig()

	s.config = &c
	s.config.Bind = "127.0.0.1:0" // Use port 0 for auto-assignment

	mc := monitoring.NewDefaultConfig()
	m, err := monitoring.New(s.T().Context(), &mc, 1)
	s.Require().NoError(err)
	s.monitoring = m

	erCfg := errorreport.NewDefaultConfig()
	er, err := errorreport.New(&erCfg)
	s.Require().NoError(err)
	s.errorReporter = er

	r, err := NewRouter(s.config, m, er)
	s.Require().NoError(err)
	s.blankRouter = r
}

func (s *ServerTestSuite) TestStartServerWithInvalidBind() {
	ctx, cancel := context.WithCancel(s.T().Context())

	// Track if cancel was called using atomic
	var cancelCalled atomic.Bool
	cancelWrapper := func() {
		cancel()
		cancelCalled.Store(true)
	}

	invalidConfig := NewDefaultConfig()
	invalidConfig.Bind = "-1.-1.-1.-1" // Invalid address

	r, err := NewRouter(&invalidConfig, s.monitoring, s.errorReporter)
	s.Require().NoError(err)

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
			config := NewDefaultConfig()
			config.CORSAllowOrigin = tt.corsAllowOrigin

			router, err := NewRouter(&config, s.monitoring, s.errorReporter)
			s.Require().NoError(err)

			wrappedHandler := router.WithCORS(s.mockHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rw := httptest.NewRecorder()

			wrappedHandler("test-req-id", s.wrapRW(rw), req)

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
			config := NewDefaultConfig()
			config.Secret = tt.secret

			router, err := NewRouter(&config, s.monitoring, s.errorReporter)
			s.Require().NoError(err)

			wrappedHandler := router.WithSecret(s.mockHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set(httpheaders.Authorization, tt.authHeader)
			}
			rw := httptest.NewRecorder()

			if serr := wrappedHandler("test-req-id", s.wrapRW(rw), req); serr != nil {
				err = serr.Err
			}

			if tt.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *ServerTestSuite) TestIntoSuccess() {
	mockHandler := func(reqID string, rw ResponseWriter, r *http.Request) *Error {
		rw.WriteHeader(http.StatusOK)
		return nil
	}

	wrappedHandler := s.blankRouter.WithReportError(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	wrappedHandler("test-req-id", s.wrapRW(rw), req)

	s.Equal(http.StatusOK, rw.Code)
}

func (s *ServerTestSuite) TestIntoWithError() {
	testError := errctx.NewTextError("test error", 0)
	mockHandler := func(reqID string, rw ResponseWriter, r *http.Request) *Error {
		return NewError(testError, "test-category")
	}

	wrappedHandler := s.blankRouter.WithReportError(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	wrappedHandler("test-req-id", s.wrapRW(rw), req)

	s.Equal(http.StatusInternalServerError, rw.Code)
	s.Equal("text/plain", rw.Header().Get(httpheaders.ContentType))
}

func (s *ServerTestSuite) TestIntoPanicWithError() {
	testError := errors.New("panic error")
	mockHandler := func(reqID string, rw ResponseWriter, r *http.Request) *Error {
		panic(testError)
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	s.NotPanics(func() {
		err := wrappedHandler("test-req-id", s.wrapRW(rw), req)
		s.Require().NotNil(err)
		s.Require().Error(err.Err, "panic error")
	})

	s.Equal(http.StatusOK, rw.Code)
}

func (s *ServerTestSuite) TestIntoPanicWithAbortHandler() {
	mockHandler := func(reqID string, rw ResponseWriter, r *http.Request) *Error {
		panic(http.ErrAbortHandler)
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	// Should re-panic with ErrAbortHandler
	s.Panics(func() {
		wrappedHandler("test-req-id", s.wrapRW(rw), req)
	})
}

func (s *ServerTestSuite) TestIntoPanicWithNonError() {
	mockHandler := func(reqID string, rw ResponseWriter, r *http.Request) *Error {
		panic("string panic")
	}

	wrappedHandler := s.blankRouter.WithPanic(mockHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	// Should re-panic with non-error panics
	s.NotPanics(func() {
		err := wrappedHandler("test-req-id", s.wrapRW(rw), req)
		s.Require().NotNil(err)
		s.Require().Error(err.Err, "string panic")
	})
}

func (s *ServerTestSuite) mockHandler(reqID string, rw ResponseWriter, r *http.Request) *Error {
	return nil
}

func (s *ServerTestSuite) wrapRW(rw http.ResponseWriter) ResponseWriter {
	return s.blankRouter.rwFactory.NewWriter(rw)
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
