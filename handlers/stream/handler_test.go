package stream

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/transport"
)

const (
	testDataPath = "../../testdata"
)

type HandlerTestSuite struct {
	suite.Suite
	handler *Handler
}

func (s *HandlerTestSuite) SetupSuite() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	// Silence logs during tests
	logrus.SetOutput(io.Discard)
}

func (s *HandlerTestSuite) TearDownSuite() {
	config.Reset()
	logrus.SetOutput(os.Stdout)
}

func (s *HandlerTestSuite) SetupTest() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	tr, err := transport.NewTransport()
	s.Require().NoError(err)

	fc := imagefetcher.NewDefaultConfig()

	fetcher, err := imagefetcher.NewFetcher(tr, fc)
	s.Require().NoError(err)

	cfg := NewDefaultConfig()

	hwc := headerwriter.NewDefaultConfig()
	hw, err := headerwriter.New(hwc)
	s.Require().NoError(err)

	h, err := New(cfg, hw, fetcher)
	s.Require().NoError(err)
	s.handler = h
}

func (s *HandlerTestSuite) readTestFile(name string) []byte {
	data, err := os.ReadFile(filepath.Join(testDataPath, name))
	s.Require().NoError(err)
	return data
}

// TestHandlerBasicRequest checks basic streaming request
func (s *HandlerTestSuite) TestHandlerBasicRequest() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheaders.ContentType, "image/png")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err := s.handler.Execute(context.Background(), req, ts.URL, "request-1", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get(httpheaders.ContentType))

	// Verify we get the original image data
	actual := rw.Body.Bytes()
	s.Require().Equal(data, actual)
}

// TestHandlerResponseHeadersPassthrough checks that original response headers are
// passed through to the client
func (s *HandlerTestSuite) TestHandlerResponseHeadersPassthrough() {
	data := s.readTestFile("test1.png")
	contentLength := len(data)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheaders.ContentType, "image/png")
		w.Header().Set(httpheaders.ContentLength, strconv.Itoa(contentLength))
		w.Header().Set(httpheaders.AcceptRanges, "bytes")
		w.Header().Set(httpheaders.Etag, "etag")
		w.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err := s.handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get(httpheaders.ContentType))
	s.Require().Equal(strconv.Itoa(contentLength), res.Header.Get(httpheaders.ContentLength))
	s.Require().Equal("bytes", res.Header.Get(httpheaders.AcceptRanges))
	s.Require().Equal("etag", res.Header.Get(httpheaders.Etag))
	s.Require().Equal("Wed, 21 Oct 2015 07:28:00 GMT", res.Header.Get(httpheaders.LastModified))
}

// TestHandlerRequestHeadersPassthrough checks that original request headers are passed through
// to the server
func (s *HandlerTestSuite) TestHandlerRequestHeadersPassthrough() {
	etag := `"test-etag-123"`
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that If-None-Match header is passed through
		s.Equal(etag, r.Header.Get(httpheaders.IfNoneMatch))
		s.Equal("gzip", r.Header.Get(httpheaders.AcceptEncoding))
		s.Equal("bytes=*", r.Header.Get(httpheaders.Range))

		w.Header().Set(httpheaders.Etag, etag)
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(httpheaders.IfNoneMatch, etag)
	req.Header.Set(httpheaders.AcceptEncoding, "gzip")
	req.Header.Set(httpheaders.Range, "bytes=*")

	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err := s.handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get(httpheaders.Etag))
}

// TestHandlerContentDisposition checks that Content-Disposition header is set correctly
func (s *HandlerTestSuite) TestHandlerContentDisposition() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheaders.ContentType, "image/png")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{
		Filename:         "custom_name",
		ReturnAttachment: true,
	}

	// Use a URL with a .png extension to help content disposition logic
	imageURL := ts.URL + "/test.png"
	err := s.handler.Execute(context.Background(), req, imageURL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
	s.Require().Contains(res.Header.Get(httpheaders.ContentDisposition), "custom_name.png")
	s.Require().Contains(res.Header.Get(httpheaders.ContentDisposition), "attachment")
}

// TestHandlerCacheControl checks that Cache-Control header is set correctly in different cases
func (s *HandlerTestSuite) TestHandlerCacheControl() {
	type testCase struct {
		name                    string
		cacheControlPassthrough bool
		setupOriginHeaders      func(http.ResponseWriter)
		timestampOffset         *time.Duration // nil for no timestamp, otherwise the offset from now
		expectedStatusCode      int
		validate                func(*testing.T, *http.Response)
	}

	// Duration variables for test cases
	var (
		oneHour          = time.Hour
		thirtyMinutes    = 30 * time.Minute
		fortyFiveMinutes = 45 * time.Minute
		twoHours         = time.Hour * 2
		oneMinuteDelta   = float64(time.Minute)
	)

	defaultTTL := 4242

	testCases := []testCase{
		{
			name:                    "Passthrough",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(w http.ResponseWriter) {
				w.Header().Set(httpheaders.CacheControl, "max-age=3600, public")
			},
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal("max-age=3600, public", res.Header.Get(httpheaders.CacheControl))
			},
		},
		// Checks that expires gets convert to cache-control
		{
			name:                    "ExpiresPassthrough",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(w http.ResponseWriter) {
				w.Header().Set(httpheaders.Expires, time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
			},
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				// When expires is converted to cache-control, the expires header should be empty
				s.Require().Empty(res.Header.Get(httpheaders.Expires))
				s.Require().InDelta(oneHour, s.maxAgeValue(res), oneMinuteDelta)
			},
		},
		// It would be set to something like default ttl
		{
			name:                    "PassthroughDisabled",
			cacheControlPassthrough: false,
			setupOriginHeaders: func(w http.ResponseWriter) {
				w.Header().Set(httpheaders.CacheControl, "max-age=3600, public")
			},
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal(s.maxAgeValue(res), time.Duration(defaultTTL)*time.Second)
			},
		},
		// When expires is set in processing options, but not present in the response
		{
			name:                    "WithProcessingOptionsExpires",
			cacheControlPassthrough: false,
			setupOriginHeaders:      func(w http.ResponseWriter) {}, // No origin headers
			timestampOffset:         &oneHour,
			expectedStatusCode:      200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().InDelta(oneHour, s.maxAgeValue(res), oneMinuteDelta)
			},
		},
		// When expires is set in processing options, and is present in the response,
		// and passthrough is enabled
		{
			name:                    "ProcessingOptionsOverridesOrigin",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(w http.ResponseWriter) {
				// Origin has a longer cache time
				w.Header().Set(httpheaders.CacheControl, "max-age=7200, public")
			},
			timestampOffset:    &thirtyMinutes,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().InDelta(thirtyMinutes, s.maxAgeValue(res), oneMinuteDelta)
			},
		},
		// When expires is not set in po, but both expires and cc are present in response,
		// and passthrough is enabled
		{
			name:                    "BothHeadersPassthroughEnabled",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(w http.ResponseWriter) {
				// Origin has both Cache-Control and Expires headers
				w.Header().Set(httpheaders.CacheControl, "max-age=1800, public")
				w.Header().Set(httpheaders.Expires, time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
			},
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				// Cache-Control should take precedence over Expires when both are present
				s.Require().InDelta(thirtyMinutes, s.maxAgeValue(res), oneMinuteDelta)
				s.Require().Empty(res.Header.Get(httpheaders.Expires))
			},
		},
		// When expires is set in PO AND both cache-control and expires are present in response,
		// and passthrough is enabled
		{
			name:                    "ProcessingOptionsOverridesBothOriginHeaders",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(w http.ResponseWriter) {
				// Origin has both Cache-Control and Expires headers with longer cache times
				w.Header().Set(httpheaders.CacheControl, "max-age=7200, public")
				w.Header().Set(httpheaders.Expires, time.Now().Add(twoHours).UTC().Format(http.TimeFormat))
			},
			timestampOffset:    &fortyFiveMinutes, // Shorter than origin headers
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().InDelta(fortyFiveMinutes, s.maxAgeValue(res), oneMinuteDelta)
				s.Require().Empty(res.Header.Get(httpheaders.Expires))
			},
		},
		// No headers set
		{
			name:                    "NoOriginHeaders",
			cacheControlPassthrough: false,
			setupOriginHeaders:      func(w http.ResponseWriter) {}, // Origin has no cache headers
			timestampOffset:         nil,
			expectedStatusCode:      200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal(s.maxAgeValue(res), time.Duration(defaultTTL)*time.Second)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			data := s.readTestFile("test1.png")

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.setupOriginHeaders(w)
				w.Header().Set(httpheaders.ContentType, "image/png")
				w.WriteHeader(200)
				w.Write(data)
			}))
			defer ts.Close()

			// Create new handler with updated config for each test
			tr, err := transport.NewTransport()
			s.Require().NoError(err)

			fc := imagefetcher.NewDefaultConfig()

			fetcher, err := imagefetcher.NewFetcher(tr, fc)
			s.Require().NoError(err)

			cfg := NewDefaultConfig()
			hwc := headerwriter.NewDefaultConfig()
			hwc.CacheControlPassthrough = tc.cacheControlPassthrough
			hwc.DefaultTTL = 4242

			hw, err := headerwriter.New(hwc)
			s.Require().NoError(err)

			handler, err := New(cfg, hw, fetcher)
			s.Require().NoError(err)

			req := httptest.NewRequest("GET", "/", nil)
			rw := httptest.NewRecorder()
			po := &options.ProcessingOptions{}

			if tc.timestampOffset != nil {
				expires := time.Now().Add(*tc.timestampOffset)
				po.Expires = &expires
			}

			err = handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
			s.Require().NoError(err)

			res := rw.Result()
			s.Require().Equal(tc.expectedStatusCode, res.StatusCode)
			tc.validate(s.T(), res)
		})
	}
}

// maxAgeValue parses max-age from cache-control
func (s *HandlerTestSuite) maxAgeValue(res *http.Response) time.Duration {
	cacheControl := res.Header.Get(httpheaders.CacheControl)
	if cacheControl == "" {
		return 0
	}
	var maxAge int
	fmt.Sscanf(cacheControl, "max-age=%d", &maxAge)
	return time.Duration(maxAge) * time.Second
}

// TestHandlerSecurityHeaders tests the security headers set by the streaming service.
func (s *HandlerTestSuite) TestHandlerSecurityHeaders() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheaders.ContentType, "image/png")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err := s.handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("script-src 'none'", res.Header.Get(httpheaders.ContentSecurityPolicy))
}

// TestHandlerErrorResponse tests the error responses from the streaming service.
func (s *HandlerTestSuite) TestHandlerErrorResponse() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("Not Found"))
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err := s.handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(404, res.StatusCode)
}

// TestHandlerCookiePassthrough tests the cookie passthrough behavior of the streaming service.
func (s *HandlerTestSuite) TestHandlerCookiePassthrough() {
	// Create new handler with updated config
	tr, err := transport.NewTransport()
	s.Require().NoError(err)

	fc := imagefetcher.NewDefaultConfig()
	fetcher, err := imagefetcher.NewFetcher(tr, fc)
	s.Require().NoError(err)

	cfg := NewDefaultConfig()
	cfg.CookiePassthrough = true

	hwc := headerwriter.NewDefaultConfig()
	hw, err := headerwriter.New(hwc)
	s.Require().NoError(err)

	handler, err := New(cfg, hw, fetcher)
	s.Require().NoError(err)

	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify cookies are passed through
		cookie, cerr := r.Cookie("test_cookie")
		if cerr == nil {
			s.Equal("test_value", cookie.Value)
		}

		w.Header().Set(httpheaders.ContentType, "image/png")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(httpheaders.Cookie, "test_cookie=test_value")
	rw := httptest.NewRecorder()
	po := &options.ProcessingOptions{}

	err = handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
	s.Require().NoError(err)

	res := rw.Result()
	s.Require().Equal(200, res.StatusCode)
}

// TestHandlerCanonicalHeader tests that the canonical header is set correctly
func (s *HandlerTestSuite) TestHandlerCanonicalHeader() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(httpheaders.ContentType, "image/png")
		w.WriteHeader(200)
		w.Write(data)
	}))
	defer ts.Close()

	for _, sc := range []bool{true, false} {
		// Create new handler with updated config
		tr, err := transport.NewTransport()
		s.Require().NoError(err)

		fc := imagefetcher.NewDefaultConfig()
		fetcher, err := imagefetcher.NewFetcher(tr, fc)
		s.Require().NoError(err)

		cfg := NewDefaultConfig()
		hwc := headerwriter.NewDefaultConfig()

		hwc.SetCanonicalHeader = sc

		hw, err := headerwriter.New(hwc)
		s.Require().NoError(err)

		handler, err := New(cfg, hw, fetcher)
		s.Require().NoError(err)

		req := httptest.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		po := &options.ProcessingOptions{}

		err = handler.Execute(context.Background(), req, ts.URL, "test-req-id", po, rw)
		s.Require().NoError(err)

		res := rw.Result()
		s.Require().Equal(200, res.StatusCode)

		if sc {
			s.Require().Contains(res.Header.Get(httpheaders.Link), fmt.Sprintf(`<%s>; rel="canonical"`, ts.URL))
		} else {
			s.Require().Empty(res.Header.Get(httpheaders.Link))
		}
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
