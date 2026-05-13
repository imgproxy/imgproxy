package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport"
	"github.com/imgproxy/imgproxy/v4/httpheaders"
	"github.com/imgproxy/imgproxy/v4/monitoring"
	"github.com/imgproxy/imgproxy/v4/server"
	"github.com/imgproxy/imgproxy/v4/testutil"
)

type ServerTestSuite struct {
	testutil.LazySuite

	config testutil.LazyObj[*server.Config]
	router testutil.LazyObj[*server.Router]
}

func (s *ServerTestSuite) SetupSuite() {
	s.config, _ = testutil.NewLazySuiteObj(s, func() (*server.Config, error) {
		c := server.NewDefaultConfig()
		c.Bind = "127.0.0.1:0" // Use port 0 for auto-assignment
		return &c, nil
	})

	s.router, _ = testutil.NewLazySuiteObj(s, func() (*server.Router, error) {
		mc := monitoring.NewDefaultConfig()
		m, err := monitoring.New(s.T().Context(), &mc, 1)
		if err != nil {
			return nil, err
		}

		erCfg := errorreport.NewDefaultConfig()
		er, err := errorreport.New(&erCfg)
		if err != nil {
			return nil, err
		}

		return server.NewRouter(s.config(), m, er)
	})
}

func (s *ServerTestSuite) SetupSubTest() {
	s.ResetLazyObjects()
}

func (s *ServerTestSuite) TestStartServerWithInvalidBind() {
	ctx, cancel := context.WithCancel(s.T().Context())

	// Track if cancel was called using atomic
	var cancelCalled atomic.Bool
	cancelWrapper := func() {
		cancel()
		cancelCalled.Store(true)
	}

	s.config().Bind = "-1.-1.-1.-1" // Invalid address

	srv, err := server.Start(cancelWrapper, s.router())

	s.Require().Error(err)
	s.Nil(srv)
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

	srv, err := server.Start(cancel, s.router())
	s.Require().NoError(err)
	s.NotNil(srv)

	// Test graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(s.T().Context(), 10*time.Second)
	defer shutdownCancel()

	// Should not panic or hang
	s.NotPanics(func() {
		srv.Shutdown(shutdownCtx)
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
			s.config().CORSAllowOrigin = tt.corsAllowOrigin

			s.router().GET("/test", s.router().WithCORS(s.mockHandler))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rw := httptest.NewRecorder()

			s.router().ServeHTTP(rw, req)

			s.Equal(tt.expectedOrigin, rw.Header().Get(httpheaders.AccessControlAllowOrigin))
			s.Equal(tt.expectedMethods, rw.Header().Get(httpheaders.AccessControlAllowMethods))
		})
	}
}

func (s *ServerTestSuite) TestWithSecret() {
	tests := []struct {
		name         string
		secret       string
		authHeader   string
		expectStatus int
	}{
		{
			name:         "ValidSecret",
			secret:       "test-secret",
			authHeader:   "Bearer test-secret",
			expectStatus: http.StatusOK,
		},
		{
			name:         "InvalidSecret",
			secret:       "foo-secret",
			authHeader:   "Bearer wrong-secret",
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "NoSecretConfigured",
			secret:       "",
			authHeader:   "",
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.config().Secret = tt.secret

			s.router().GET("/test", s.router().WithReportError(s.router().WithSecret(s.mockHandler)))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set(httpheaders.Authorization, tt.authHeader)
			}
			rw := httptest.NewRecorder()

			s.router().ServeHTTP(rw, req)

			s.Equal(tt.expectStatus, rw.Code)
		})
	}
}

func (s *ServerTestSuite) TestIntoSuccess() {
	mockHandler := func(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
		rw.WriteHeader(http.StatusOK)
		return nil
	}

	s.router().GET("/test", s.router().WithReportError(mockHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	s.router().ServeHTTP(rw, req)

	s.Equal(http.StatusOK, rw.Code)
}

func (s *ServerTestSuite) TestIntoWithError() {
	testError := errctx.NewTextError("test error", 0)
	mockHandler := func(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
		return server.NewError(testError, "test-category")
	}

	s.router().GET("/test", s.router().WithReportError(mockHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	s.router().ServeHTTP(rw, req)

	s.Equal(http.StatusInternalServerError, rw.Code)
	s.Equal("text/plain", rw.Header().Get(httpheaders.ContentType))
}

func (s *ServerTestSuite) TestIntoPanicWithError() {
	testError := errors.New("panic error")
	mockHandler := func(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
		panic(testError)
	}

	s.router().GET("/test", s.router().WithReportError(s.router().WithPanic(mockHandler)))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	s.NotPanics(func() {
		s.router().ServeHTTP(rw, req)
	})

	s.Equal(http.StatusInternalServerError, rw.Code)
}

func (s *ServerTestSuite) TestIntoPanicWithAbortHandler() {
	mockHandler := func(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
		panic(http.ErrAbortHandler)
	}

	s.router().GET("/test", s.router().WithPanic(mockHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	// Should re-panic with ErrAbortHandler
	s.Panics(func() {
		s.router().ServeHTTP(rw, req)
	})
}

func (s *ServerTestSuite) TestIntoPanicWithNonError() {
	mockHandler := func(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
		panic("string panic")
	}

	s.router().GET("/test", s.router().WithReportError(s.router().WithPanic(mockHandler)))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rw := httptest.NewRecorder()

	s.NotPanics(func() {
		s.router().ServeHTTP(rw, req)
	})

	s.Equal(http.StatusInternalServerError, rw.Code)
}

func (s *ServerTestSuite) mockHandler(reqID string, rw server.ResponseWriter, r *http.Request) *server.Error {
	return nil
}

func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}
