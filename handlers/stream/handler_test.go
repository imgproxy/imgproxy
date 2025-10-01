package stream

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/server/responsewriter"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type HandlerTestSuite struct {
	testutil.LazySuite

	testData *testutil.TestDataProvider

	rwConf    testutil.LazyObj[*responsewriter.Config]
	rwFactory testutil.LazyObj[*responsewriter.Factory]

	cookieConf testutil.LazyObj[*cookies.Config]
	cookies    testutil.LazyObj[*cookies.Cookies]

	config  testutil.LazyObj[*Config]
	handler testutil.LazyObj[*Handler]

	testServer testutil.LazyTestServer
}

func (s *HandlerTestSuite) SetupSuite() {
	config.Reset()

	s.testData = testutil.NewTestDataProvider(s.T)

	s.rwConf, _ = testutil.NewLazySuiteObj(
		s,
		func() (*responsewriter.Config, error) {
			c := responsewriter.NewDefaultConfig()
			return &c, nil
		},
	)

	s.rwFactory, _ = testutil.NewLazySuiteObj(
		s,
		func() (*responsewriter.Factory, error) {
			return responsewriter.NewFactory(s.rwConf())
		},
	)

	s.cookieConf, _ = testutil.NewLazySuiteObj(
		s,
		func() (*cookies.Config, error) {
			c := cookies.NewDefaultConfig()
			return &c, nil
		},
	)

	s.cookies, _ = testutil.NewLazySuiteObj(
		s,
		func() (*cookies.Cookies, error) {
			return cookies.New(s.cookieConf())
		},
	)

	s.config, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Config, error) {
			c := NewDefaultConfig()
			return &c, nil
		},
	)

	s.handler, _ = testutil.NewLazySuiteObj(
		s,
		func() (*Handler, error) {
			fc := fetcher.NewDefaultConfig()
			fc.Transport.HTTP.AllowLoopbackSourceAddresses = true

			fetcher, err := fetcher.New(&fc)
			s.Require().NoError(err)

			return New(s.config(), fetcher, s.cookies())
		},
	)

	s.testServer, _ = testutil.NewLazySuiteTestServer(s)

	// Silence logs during tests
	logger.Mute()
}

func (s *HandlerTestSuite) TearDownSuite() {
	logger.Unmute()
}

func (s *HandlerTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

func (s *HandlerTestSuite) execute(
	imageURL string,
	header http.Header,
	o *options.Options,
) *http.Response {
	imageURL = s.testServer().URL() + imageURL
	req := httptest.NewRequest("GET", "/", nil)
	httpheaders.CopyAll(header, req.Header, true)

	ctx := s.T().Context()
	rw := httptest.NewRecorder()
	rww := s.rwFactory().NewWriter(rw)

	err := s.handler().Execute(ctx, req, imageURL, "test-req-id", o, rww)
	s.Require().NoError(err)

	return rw.Result()
}

// TestHandlerBasicRequest checks basic streaming request
func (s *HandlerTestSuite) TestHandlerBasicRequest() {
	data := s.testData.Read("test1.png")

	s.testServer().SetHeaders(httpheaders.ContentType, "image/png").SetBody(data)

	res := s.execute("", nil, options.New())

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get(httpheaders.ContentType))

	// Verify we get the original image data
	actual, err := io.ReadAll(res.Body)
	s.Require().NoError(err)
	s.Require().Equal(data, actual)
}

// TestHandlerResponseHeadersPassthrough checks that original response headers are
// passed through to the client
func (s *HandlerTestSuite) TestHandlerResponseHeadersPassthrough() {
	data := s.testData.Read("test1.png")
	contentLength := len(data)

	s.testServer().SetHeaders(
		httpheaders.ContentType, "image/png",
		httpheaders.ContentLength, strconv.Itoa(contentLength),
		httpheaders.AcceptRanges, "bytes",
		httpheaders.Etag, "etag",
		httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT",
	).SetBody(data)

	res := s.execute("", nil, options.New())

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
	data := s.testData.Read("test1.png")

	s.testServer().
		SetBody(data).
		SetHeaders(httpheaders.Etag, etag).
		SetHook(func(r *http.Request, rw http.ResponseWriter) {
			// Verify that If-None-Match header is passed through
			s.Equal(etag, r.Header.Get(httpheaders.IfNoneMatch))
			s.Equal("gzip", r.Header.Get(httpheaders.AcceptEncoding))
			s.Equal("bytes=*", r.Header.Get(httpheaders.Range))
		})

	h := make(http.Header)
	h.Set(httpheaders.IfNoneMatch, etag)
	h.Set(httpheaders.AcceptEncoding, "gzip")
	h.Set(httpheaders.Range, "bytes=*")

	res := s.execute("", h, options.New())

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get(httpheaders.Etag))
}

// TestHandlerContentDisposition checks that Content-Disposition header is set correctly
func (s *HandlerTestSuite) TestHandlerContentDisposition() {
	data := s.testData.Read("test1.png")

	s.testServer().SetHeaders(httpheaders.ContentType, "image/png").SetBody(data)

	o := options.New()
	o.Set(keys.Filename, "custom_name")
	o.Set(keys.ReturnAttachment, true)

	// Use a URL with a .png extension to help content disposition logic
	res := s.execute("/test.png", nil, o)

	s.Require().Equal(200, res.StatusCode)
	s.Require().Contains(res.Header.Get(httpheaders.ContentDisposition), "custom_name.png")
	s.Require().Contains(res.Header.Get(httpheaders.ContentDisposition), "attachment")
}

// TestHandlerCacheControl checks that Cache-Control header is set correctly in different cases
func (s *HandlerTestSuite) TestHandlerCacheControl() {
	type testCase struct {
		name                    string
		cacheControlPassthrough bool
		setupOriginHeaders      func()
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
			setupOriginHeaders: func() {
				s.testServer().SetHeaders(httpheaders.CacheControl, "max-age=3600, public")
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
			setupOriginHeaders: func() {
				s.testServer().SetHeaders(httpheaders.Expires, time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
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
			setupOriginHeaders: func() {
				s.testServer().SetHeaders(httpheaders.CacheControl, "max-age=3600, public")
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
			setupOriginHeaders: func() {
				// Origin has a longer cache time
				s.testServer().SetHeaders(httpheaders.CacheControl, "max-age=7200, public")
			},
			timestampOffset:    &thirtyMinutes,
			expectedStatusCode: 200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().InDelta(thirtyMinutes, s.maxAgeValue(res), oneMinuteDelta)
			},
		},
		// When expires is not set in o, but both expires and cc are present in response,
		// and passthrough is enabled
		{
			name:                    "BothHeadersPassthroughEnabled",
			cacheControlPassthrough: true,
			setupOriginHeaders: func() {
				// Origin has both Cache-Control and Expires headers
				s.testServer().SetHeaders(httpheaders.CacheControl, "max-age=1800, public")
				s.testServer().SetHeaders(httpheaders.Expires, time.Now().Add(oneHour).UTC().Format(http.TimeFormat))
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
			setupOriginHeaders: func() {
				// Origin has both Cache-Control and Expires headers with longer cache times
				s.testServer().SetHeaders(httpheaders.CacheControl, "max-age=7200, public")
				s.testServer().SetHeaders(httpheaders.Expires, time.Now().Add(twoHours).UTC().Format(http.TimeFormat))
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
			timestampOffset:         nil,
			expectedStatusCode:      200,
			validate: func(t *testing.T, res *http.Response) {
				s.Require().Equal(s.maxAgeValue(res), time.Duration(defaultTTL)*time.Second)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			data := s.testData.Read("test1.png")

			if tc.setupOriginHeaders != nil {
				tc.setupOriginHeaders()
			}

			s.testServer().SetHeaders(httpheaders.ContentType, "image/png").SetBody(data)

			s.rwConf().CacheControlPassthrough = tc.cacheControlPassthrough
			s.rwConf().DefaultTTL = 4242

			o := options.New()

			if tc.timestampOffset != nil {
				o.Set(keys.Expires, time.Now().Add(*tc.timestampOffset))
			}

			res := s.execute("", nil, o)
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
	data := s.testData.Read("test1.png")

	s.testServer().SetHeaders(httpheaders.ContentType, "image/png").SetBody(data)

	res := s.execute("", nil, options.New())

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("script-src 'none'", res.Header.Get(httpheaders.ContentSecurityPolicy))
}

// TestHandlerErrorResponse tests the error responses from the streaming service.
func (s *HandlerTestSuite) TestHandlerErrorResponse() {
	s.testServer().SetStatusCode(http.StatusNotFound).SetBody([]byte("Not Found"))

	res := s.execute("", nil, options.New())

	s.Require().Equal(http.StatusNotFound, res.StatusCode)
}

// TestHandlerCookiePassthrough tests the cookie passthrough behavior of the streaming service.
func (s *HandlerTestSuite) TestHandlerCookiePassthrough() {
	s.cookieConf().CookiePassthrough = true

	data := s.testData.Read("test1.png")

	s.testServer().
		SetHeaders(httpheaders.Cookie, "test_cookie=test_value").
		SetHook(func(r *http.Request, rw http.ResponseWriter) {
			// Verify cookies are passed through
			cookie, cerr := r.Cookie("test_cookie")
			if cerr == nil {
				s.Equal("test_value", cookie.Value)
			}
		}).SetBody(data)

	h := make(http.Header)
	h.Set(httpheaders.Cookie, "test_cookie=test_value")

	res := s.execute("", h, options.New())

	s.Require().Equal(200, res.StatusCode)
}

// TestHandlerCanonicalHeader tests that the canonical header is set correctly
func (s *HandlerTestSuite) TestHandlerCanonicalHeader() {
	data := s.testData.Read("test1.png")

	s.testServer().SetHeaders(httpheaders.ContentType, "image/png").SetBody(data)

	for _, sc := range []bool{true, false} {
		s.rwConf().SetCanonicalHeader = sc

		res := s.execute("", nil, options.New())

		s.Require().Equal(200, res.StatusCode)

		if sc {
			s.Require().Contains(res.Header.Get(httpheaders.Link), fmt.Sprintf(`<%s>; rel="canonical"`, s.testServer().URL()))
		} else {
			s.Require().Empty(res.Header.Get(httpheaders.Link))
		}
	}
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
