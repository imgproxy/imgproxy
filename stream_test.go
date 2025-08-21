package main

import (
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
	"github.com/imgproxy/imgproxy/v3/server"
)

type StreamTestSuite struct {
	suite.Suite

	router *server.Router
}

func (s *StreamTestSuite) SetupSuite() {
	config.Reset()

	wd, err := os.Getwd()
	s.Require().NoError(err)

	s.T().Setenv("IMGPROXY_LOCAL_FILESYSTEM_ROOT", filepath.Join(wd, "/testdata"))
	s.T().Setenv("IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT", "0")

	err = initialize()
	s.Require().NoError(err)

	logrus.SetOutput(io.Discard)

	s.router = buildRouter(server.NewRouter(server.NewConfigFromEnv()))
}

func (s *StreamTestSuite) TeardownSuite() {
	shutdown()
	logrus.SetOutput(os.Stdout)
}

func (s *StreamTestSuite) SetupTest() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true
}

func (s *StreamTestSuite) send(path string, header http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rw := httptest.NewRecorder()

	req.Header = header

	s.router.ServeHTTP(rw, req)

	return rw
}

func (s *StreamTestSuite) readTestFile(name string) []byte {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(wd, "testdata", name))
	s.Require().NoError(err)

	return data
}

// TestStreamBasicRequest checks basic streaming request
func (s *StreamTestSuite) TestStreamBasicRequest() {
	rw := s.send("/unsafe/raw:1/plain/local:///test1.png", nil)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get("Content-Type"))

	// Verify we get the original image data without processing
	expected := s.readTestFile("test1.png")
	actual := rw.Body.Bytes()
	s.Require().Equal(expected, actual)
}

// TestStreamResponseHeadersPassthrough checks that original response headers are
// passed through to the client
func (s *StreamTestSuite) TestStreamResponseHeadersPassthrough() {
	data := s.readTestFile("test1.png")
	contentLength := len(data)

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "image/png")
		rw.Header().Set("Content-Length", strconv.Itoa(contentLength))
		rw.Header().Set("Accept-Ranges", "bytes")
		rw.Header().Set("ETag", "etag")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, nil)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get("Content-Type"))
	s.Require().Equal(strconv.Itoa(contentLength), res.Header.Get("Content-Length"))
	s.Require().Equal("bytes", res.Header.Get("Accept-Ranges"))
	s.Require().Equal("etag", res.Header.Get("ETag"))
}

// TestStreamLastModifiedPassthrough checks that Last-Modified header is passed through from the
// server to the response regardless of config.LastModifiedEnabled setting
func (s *StreamTestSuite) TestStreamLastModifiedPassthrough() {
	config.LastModifiedEnabled = false
	data := s.readTestFile("test1.png")
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Last-Modified", lastModified)
		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, nil)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(lastModified, res.Header.Get("Last-Modified"))
}

// TestStreamRequestHeadersPassthrough checks that original request headers are passed through
// to the server
func (s *StreamTestSuite) TestStreamRequestHeadersPassthrough() {
	etag := `"test-etag-123"`
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Verify that If-None-Match header is passed through
		s.Equal(etag, r.Header.Get("If-None-Match"))
		s.Equal("test", r.Header.Get("If-Modified-Since"))
		s.Equal("gzip", r.Header.Get("Accept-Encoding"))
		s.Equal("bytes=*", r.Header.Get("Range"))

		rw.Header().Set("ETag", etag)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", etag)
	header.Set("If-Modified-Since", "test")
	header.Set("Accept-Encoding", "gzip")
	header.Set("Range", "bytes=*")

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get("ETag"))
}

// TestStreamContentDisposition checks that Content-Disposition header is set correctly
func (s *StreamTestSuite) TestStreamContentDisposition() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	// Test with attachment
	rw := s.send("/unsafe/raw:1/fn:custom_name/att:1/plain/"+ts.URL, nil)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Contains(res.Header.Get("Content-Disposition"), "custom_name.png")
	s.Require().Contains(res.Header.Get("Content-Disposition"), "attachment")
}

// TestStreamCacheControl checks that Cache-Control header is set correctly in different cases
func (s *StreamTestSuite) TestStreamCacheControl() {
	type testCase struct {
		name                    string
		cacheControlPassthrough bool
		setupOriginHeaders      func(http.ResponseWriter)
		urlPath                 string
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

	// Set this explicitly for testing purposes
	config.TTL = 4242

	testCases := []testCase{
		{
			name:                    "Passthrough",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(rw http.ResponseWriter) {
				rw.Header().Set("Cache-Control", "max-age=3600, public")
			},
			urlPath:            "/unsafe/raw:1/plain/%s",
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal("max-age=3600, public", res.Header.Get("Cache-Control"))
			},
		},
		// Checks that expires gets convert to cache-control
		{
			name:                    "ExpiresPassthrough",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(rw http.ResponseWriter) {
				rw.Header().Set("Expires", time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
			},
			urlPath:            "/unsafe/raw:1/plain/%s",
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				// When expires is converted to cache-control, the expires header should be empty
				s.Require().Empty(res.Header.Get("Expires"))
				s.Require().InDelta(oneHour, s.maxAgeValue(res), oneMinuteDelta)
			},
		},
		// It would be set to something like default ttl
		{
			name:                    "PassthroughDisabled",
			cacheControlPassthrough: false,
			setupOriginHeaders: func(rw http.ResponseWriter) {
				rw.Header().Set("Cache-Control", "max-age=3600, public")
			},
			urlPath:            "/unsafe/raw:1/plain/%s",
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal(s.maxAgeValue(res), time.Duration(config.TTL)*time.Second)
			},
		},
		// When expires is set in processing options, but not present in the response
		{
			name:                    "WithProcessingOptionsExpires",
			cacheControlPassthrough: false,
			setupOriginHeaders:      func(rw http.ResponseWriter) {}, // No origin headers
			urlPath:                 "/unsafe/raw:1/exp:%d/plain/%s",
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
			setupOriginHeaders: func(rw http.ResponseWriter) {
				// Origin has a longer cache time
				rw.Header().Set("Cache-Control", "max-age=7200, public")
			},
			urlPath:            "/unsafe/raw:1/exp:%d/plain/%s",
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
			setupOriginHeaders: func(rw http.ResponseWriter) {
				// Origin has both Cache-Control and Expires headers
				rw.Header().Set("Cache-Control", "max-age=1800, public")
				rw.Header().Set("Expires", time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
			},
			urlPath:            "/unsafe/raw:1/plain/%s",
			timestampOffset:    nil,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				// Cache-Control should take precedence over Expires when both are present
				s.Require().InDelta(thirtyMinutes, s.maxAgeValue(res), oneMinuteDelta)
				s.Require().Empty(res.Header.Get("Expires"))
			},
		},
		// When expires is set in PO AND both cache-control and expires are present in response,
		// and passthrough is enabled
		{
			name:                    "ProcessingOptionsOverridesBothOriginHeaders",
			cacheControlPassthrough: true,
			setupOriginHeaders: func(rw http.ResponseWriter) {
				// Origin has both Cache-Control and Expires headers with longer cache times
				rw.Header().Set("Cache-Control", "max-age=7200, public")
				rw.Header().Set("Expires", time.Now().Add(twoHours).UTC().Format(http.TimeFormat))
			},
			urlPath:            "/unsafe/raw:1/exp:%d/plain/%s",
			timestampOffset:    &fortyFiveMinutes, // Shorter than origin headers
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().InDelta(fortyFiveMinutes, s.maxAgeValue(res), oneMinuteDelta)
				s.Require().Empty(res.Header.Get("Expires"))
			},
		},
		// No headers set
		{
			name:                    "NoOriginHeaders",
			cacheControlPassthrough: false,
			setupOriginHeaders:      func(rw http.ResponseWriter) {}, // Origin has no cache headers
			urlPath:                 "/unsafe/raw:1/plain/%s",
			timestampOffset:         nil,
			expectedStatusCode:      200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal(s.maxAgeValue(res), time.Duration(config.TTL)*time.Second)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			config.CacheControlPassthrough = tc.cacheControlPassthrough

			data := s.readTestFile("test1.png")

			ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				tc.setupOriginHeaders(rw)
				rw.Header().Set("Content-Type", "image/png")
				rw.WriteHeader(200)
				rw.Write(data)
			}))
			defer ts.Close()

			var url string
			if tc.timestampOffset != nil {
				timestamp := time.Now().Add(*tc.timestampOffset).Unix()
				url = fmt.Sprintf(tc.urlPath, timestamp, ts.URL)
			} else {
				url = fmt.Sprintf(tc.urlPath, ts.URL)
			}

			rw := s.send(url, nil)
			res := rw.Result()

			s.Require().Equal(tc.expectedStatusCode, res.StatusCode)
			tc.validate(s.T(), res)
		})
	}
}

// maxAgeValue parses max-age from cache-control
func (s *StreamTestSuite) maxAgeValue(res *http.Response) time.Duration {
	cacheControl := res.Header.Get("Cache-Control")
	if cacheControl == "" {
		return 0
	}
	var maxAge int
	fmt.Sscanf(cacheControl, "max-age=%d", &maxAge)
	return time.Duration(maxAge) * time.Second
}

// TestStreamSecurityHeaders tests the security headers set by the streaming service.
func (s *StreamTestSuite) TestStreamSecurityHeaders() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, nil)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("script-src 'none'", res.Header.Get("Content-Security-Policy"))
}

// TestStreamErrorResponse tests the error responses from the streaming service.
func (s *StreamTestSuite) TestStreamErrorResponse() {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(404)
		rw.Write([]byte("Not Found"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, nil)
	res := rw.Result()

	s.Require().Equal(404, res.StatusCode)
}

// TestStreamCookiePassthrough tests the cookie passthrough behavior of the streaming service.
func (s *StreamTestSuite) TestStreamCookiePassthrough() {
	config.CookiePassthrough = true

	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Verify cookies are passed through
		cookie, err := r.Cookie("test_cookie")
		if err == nil {
			s.Equal("test_value", cookie.Value)
		}

		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("Cookie", "test_cookie=test_value")

	rw := s.send("/unsafe/raw:1/plain/"+ts.URL, header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
}

// TestStreamCanonicalHeader tests that the canonical header is set correctly
func (s *StreamTestSuite) TestStreamCanonicalHeader() {
	data := s.readTestFile("test1.png")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "image/png")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	for _, sc := range []bool{true, false} {
		config.SetCanonicalHeader = sc

		rw := s.send("/unsafe/raw:1/plain/"+ts.URL, nil)
		res := rw.Result()

		s.Require().Equal(200, res.StatusCode)

		if sc {
			s.Require().Contains(res.Header.Get("Link"), fmt.Sprintf(`<%s>; rel="canonical"`, ts.URL))
		} else {
			s.Require().Empty(res.Header.Get("Link"))
		}
	}
}

func TestStream(t *testing.T) {
	suite.Run(t, new(StreamTestSuite))
}
