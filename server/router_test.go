package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/errorreport"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/monitoring"
)

type RouterTestSuite struct {
	suite.Suite
	router *Router
}

func (s *RouterTestSuite) SetupTest() {
	c := NewDefaultConfig()

	mc := monitoring.NewDefaultConfig()
	m, err := monitoring.New(s.T().Context(), &mc, 1)
	s.Require().NoError(err)

	erCfg := errorreport.NewDefaultConfig()
	er, err := errorreport.New(&erCfg)
	s.Require().NoError(err)

	c.PathPrefix = "/api"
	r, err := NewRouter(&c, m, er)
	s.Require().NoError(err)

	s.router = r
}

// TestHTTPMethods tests route methods registration and HTTP requests
func (s *RouterTestSuite) TestHTTPMethods() {
	var capturedMethod string
	var capturedPath string

	getHandler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		capturedMethod = req.Method
		capturedPath = req.URL.Path
		rw.WriteHeader(200)
		rw.Write([]byte("GET response"))
		return nil
	}

	optionsHandler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		capturedMethod = req.Method
		capturedPath = req.URL.Path
		rw.WriteHeader(200)
		rw.Write([]byte("OPTIONS response"))
		return nil
	}

	headHandler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		capturedMethod = req.Method
		capturedPath = req.URL.Path
		rw.WriteHeader(200)
		return nil
	}

	// Register routes with different configurations
	s.router.GET("/get-test", getHandler)              // exact match
	s.router.OPTIONS("/options-test*", optionsHandler) // prefix match
	s.router.HEAD("/head-test", headHandler)           // exact match

	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		expectedBody  string
		expectedPath  string
	}{
		{
			name:          "GET",
			requestMethod: http.MethodGet,
			requestPath:   "/api/get-test",
			expectedBody:  "GET response",
			expectedPath:  "/api/get-test",
		},
		{
			name:          "OPTIONS",
			requestMethod: http.MethodOptions,
			requestPath:   "/api/options-test",
			expectedBody:  "OPTIONS response",
			expectedPath:  "/api/options-test",
		},
		{
			name:          "OPTIONSPrefixed",
			requestMethod: http.MethodOptions,
			requestPath:   "/api/options-test/sub",
			expectedBody:  "OPTIONS response",
			expectedPath:  "/api/options-test/sub",
		},
		{
			name:          "HEAD",
			requestMethod: http.MethodHead,
			requestPath:   "/api/head-test",
			expectedBody:  "",
			expectedPath:  "/api/head-test",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			rw := httptest.NewRecorder()

			s.router.ServeHTTP(rw, req)

			s.Require().Equal(tt.expectedBody, rw.Body.String())
			s.Require().Equal(tt.requestMethod, capturedMethod)
			s.Require().Equal(tt.expectedPath, capturedPath)
		})
	}
}

// TestMiddlewareOrder checks middleware ordering and functionality
func (s *RouterTestSuite) TestMiddlewareOrder() {
	var order []string

	middleware1 := func(next RouteHandler) RouteHandler {
		return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
			order = append(order, "middleware1")
			return next(reqID, rw, req)
		}
	}

	middleware2 := func(next RouteHandler) RouteHandler {
		return func(reqID string, rw ResponseWriter, req *http.Request) *Error {
			order = append(order, "middleware2")
			return next(reqID, rw, req)
		}
	}

	handler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		order = append(order, "handler")
		rw.WriteHeader(200)
		return nil
	}

	s.router.GET("/test", handler, middleware2, middleware1)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rw := httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	// Middleware should execute in the order they are passed (first added first)
	s.Require().Equal([]string{"middleware1", "middleware2", "handler"}, order)
}

// TestServeHTTP tests ServeHTTP method
func (s *RouterTestSuite) TestServeHTTP() {
	handler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		rw.Header().Set("Custom-Header", "test-value")
		rw.WriteHeader(200)
		rw.Write([]byte("success"))
		return nil
	}

	s.router.GET("/test", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rw := httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	s.Require().Equal(200, rw.Code)
	s.Require().Equal("success", rw.Body.String())
	s.Require().Equal("test-value", rw.Header().Get("Custom-Header"))
	s.Require().Equal(defaultServerName, rw.Header().Get(httpheaders.Server))
	s.Require().NotEmpty(rw.Header().Get(httpheaders.XRequestID))
}

// TestRequestID checks request ID generation and validation
func (s *RouterTestSuite) TestRequestID() {
	handler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		rw.WriteHeader(200)
		return nil
	}

	s.router.GET("/test", handler)

	// Test request ID passthrough (if present)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(httpheaders.XRequestID, "valid-id-123")
	rw := httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	s.Require().Equal("valid-id-123", rw.Header().Get(httpheaders.XRequestID))

	// Test invalid request ID (should generate a new one)
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(httpheaders.XRequestID, "invalid id with spaces!")
	rw = httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	generatedID := rw.Header().Get(httpheaders.XRequestID)
	s.Require().NotEqual("invalid id with spaces!", generatedID)
	s.Require().NotEmpty(generatedID)

	// Test no request ID (should generate a new one)
	req = httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rw = httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	generatedID = rw.Header().Get(httpheaders.XRequestID)
	s.Require().NotEmpty(generatedID)
	s.Require().Regexp(`^[A-Za-z0-9_\-]+$`, generatedID)
}

// TestLambdaRequestIDExtraction checks AWS lambda request id extraction
func (s *RouterTestSuite) TestLambdaRequestIDExtraction() {
	handler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		rw.WriteHeader(200)
		return nil
	}

	s.router.GET("/test", handler)

	// Test with valid Lambda context
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(httpheaders.XAmznRequestContextHeader, `{"requestId":"lambda-req-123"}`)
	rw := httptest.NewRecorder()

	s.router.ServeHTTP(rw, req)

	s.Require().Equal("lambda-req-123", rw.Header().Get(httpheaders.XRequestID))
}

// Test IP address handling
func (s *RouterTestSuite) TestReplaceIP() {
	var capturedRemoteAddr string
	handler := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		capturedRemoteAddr = req.RemoteAddr
		rw.WriteHeader(200)
		return nil
	}

	s.router.GET("/test", handler)

	tests := []struct {
		name         string
		originalAddr string
		headers      map[string]string
		expectedAddr string
	}{
		{
			name:         "CFConnectingIP",
			originalAddr: "original:8080",
			headers: map[string]string{
				httpheaders.CFConnectingIP: "1.2.3.4",
			},
			expectedAddr: "1.2.3.4:8080",
		},
		{
			name:         "XForwardedForMulti",
			originalAddr: "original:8080",
			headers: map[string]string{
				httpheaders.XForwardedFor: "5.6.7.8, 9.10.11.12",
			},
			expectedAddr: "5.6.7.8:8080",
		},
		{
			name:         "XForwardedForSingle",
			originalAddr: "original:8080",
			headers: map[string]string{
				httpheaders.XForwardedFor: "13.14.15.16",
			},
			expectedAddr: "13.14.15.16:8080",
		},
		{
			name:         "XRealIP",
			originalAddr: "original:8080",
			headers: map[string]string{
				httpheaders.XRealIP: "17.18.19.20",
			},
			expectedAddr: "17.18.19.20:8080",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.RemoteAddr = tt.originalAddr

			for header, value := range tt.headers {
				req.Header.Set(header, value)
			}

			rw := httptest.NewRecorder()

			s.router.ServeHTTP(rw, req)

			s.Require().Equal(tt.expectedAddr, capturedRemoteAddr)
		})
	}
}

// TestRouteOrder checks exact/non-exact insertion order
func (s *RouterTestSuite) TestRouteOrder() {
	h := func(reqID string, rw ResponseWriter, req *http.Request) *Error {
		return nil
	}

	s.router.GET("/test*", h)
	s.router.GET("/test/path", h)
	s.router.GET("/test/path/nested", h)

	s.Require().Equal("/api/test/path", s.router.routes[0].path)
	s.Require().Equal("/api/test/path/nested", s.router.routes[1].path)
	s.Require().Equal("/api/test", s.router.routes[2].path)
}

func TestRouterSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}
