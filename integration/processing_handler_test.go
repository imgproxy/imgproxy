package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/svg"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

// ProcessingHandlerTestSuite is a test suite for testing image processing handler
type ProcessingHandlerTestSuite struct {
	Suite
}

func (s *ProcessingHandlerTestSuite) SetupTest() {
	config.Reset() // We reset config only at the start of each test
	s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = true
}

func (s *ProcessingHandlerTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationFailure() {
	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}

	tt := []struct {
		name       string
		url        string
		statusCode int
	}{
		{
			name:       "NoSignature",
			url:        "/unsafe/rs:fill:4:4/plain/local:///test1.png",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "BadSignature",
			url:        "/bad-signature/rs:fill:4:4/plain/local:///test1.png",
			statusCode: http.StatusForbidden,
		},
		{
			name:       "ValidSignature",
			url:        "/My9d3xq_PYpVHsPrCyww0Kh1w5KZeZhIlWhsa4az1TI/rs:fill:4:4/plain/local:///test1.png",
			statusCode: http.StatusOK,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			res := s.GET(tc.url)
			s.Require().Equal(tc.statusCode, res.StatusCode)
		})
	}
}

func (s *ProcessingHandlerTestSuite) TestSourceValidation() {
	imagedata.RedirectAllRequestsTo("local:///test1.png")
	defer imagedata.StopRedirectingRequests()

	tt := []struct {
		name           string
		allowedSources []string
		requestPath    string
		expectedError  bool
	}{
		{
			name:           "match http URL without wildcard",
			allowedSources: []string{"local://", "http://images.dev/"},
			requestPath:    "/unsafe/plain/http://images.dev/lorem/ipsum.jpg",
		},
		{
			name:           "match http URL with wildcard in hostname single level",
			allowedSources: []string{"local://", "http://*.mycdn.dev/"},
			requestPath:    "/unsafe/plain/http://a-1.mycdn.dev/lorem/ipsum.jpg",
		},
		{
			name:           "match http URL with wildcard in hostname multiple levels",
			allowedSources: []string{"local://", "http://*.mycdn.dev/"},
			requestPath:    "/unsafe/plain/http://a-1.b-2.mycdn.dev/lorem/ipsum.jpg",
		},
		{
			name:           "no match s3 URL with allowed local and http URLs",
			allowedSources: []string{"local://", "http://images.dev/"},
			requestPath:    "/unsafe/plain/s3://images/lorem/ipsum.jpg",
			expectedError:  true,
		},
		{
			name:           "no match http URL with wildcard in hostname including slash",
			allowedSources: []string{"local://", "http://*.mycdn.dev/"},
			requestPath:    "/unsafe/plain/http://other.dev/.mycdn.dev/lorem/ipsum.jpg",
			expectedError:  true,
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			config.AllowedSources = make([]*regexp.Regexp, len(tc.allowedSources))
			for i, pattern := range tc.allowedSources {
				config.AllowedSources[i] = configurators.RegexpFromPattern(pattern)
			}

			res := s.GET(tc.requestPath)

			if tc.expectedError {
				s.Require().Equal(http.StatusNotFound, res.StatusCode)
			} else {
				s.Require().Equal(http.StatusOK, res.StatusCode)
			}
		})
	}
}

func (s *ProcessingHandlerTestSuite) TestSourceNetworkValidation() {
	data := s.TestData.Read("test1.png")

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer server.Close()

	url := fmt.Sprintf("/unsafe/rs:fill:4:4/plain/%s/test1.png", server.URL)

	// We wrap this in a subtest to reset s.router()
	s.Run("AllowLoopbackSourceAddressesTrue", func() {
		s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = true
		res := s.GET(url)
		s.Require().Equal(http.StatusOK, res.StatusCode)
	})

	s.Run("AllowLoopbackSourceAddressesFalse", func() {
		s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = false
		res := s.GET(url)
		s.Require().Equal(http.StatusNotFound, res.StatusCode)
	})
}

func (s *ProcessingHandlerTestSuite) TestSourceFormatNotSupported() {
	vips.DisableLoadSupport(imagetype.PNG)
	defer vips.ResetLoadSupport()

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestResultingFormatNotSupported() {
	vips.DisableSaveSupport(imagetype.PNG)
	defer vips.ResetSaveSupport()

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingConfig() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingPO() {
	res := s.GET("/unsafe/rs:fill:4:4/skp:png/plain/local:///test1.png")

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSameFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingDifferentFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@jpg")

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().False(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSVG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.svg")

	s.Require().Equal(http.StatusOK, res.StatusCode)

	c := fetcher.NewDefaultConfig()
	f, err := fetcher.New(&c)
	s.Require().NoError(err)

	idf := imagedata.NewFactory(f)

	data, err := idf.NewFromBytes(s.TestData.Read("test1.svg"))
	s.Require().NoError(err)

	expected, err := svg.Sanitize(data)
	s.Require().NoError(err)

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestNotSkipProcessingSVGToJPG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.svg@jpg")

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().False(s.TestData.FileEqualsToReader("test1.svg", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestErrorSavingToSVG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@svg")

	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughCacheControl() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.CacheControl, "max-age=1234, public")
		rw.Header().Set(httpheaders.Expires, time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("max-age=1234, public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughExpires() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.Expires, time.Now().Add(1239*time.Second).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	// Use regex to allow some delay
	s.Require().Regexp("max-age=123[0-9], public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughDisabled() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = false

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.CacheControl, "max-age=1234, public")
		rw.Header().Set(httpheaders.Expires, time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	s.Require().NotEqual("max-age=1234, public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestETagDisabled() {
	s.Config().Handlers.Processing.ETagEnabled = false

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")

	s.Require().Equal(200, res.StatusCode)
	s.Require().Empty(res.Header.Get(httpheaders.Etag))
}

func (s *ProcessingHandlerTestSuite) TestETagDataMatch() {
	s.Config().Handlers.Processing.ETagEnabled = true

	etag := `"loremipsumdolor"`

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.NotEmpty(r.Header.Get(httpheaders.IfNoneMatch))

		rw.Header().Set(httpheaders.Etag, etag)
		rw.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set(httpheaders.IfNoneMatch, etag)

	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(304, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get(httpheaders.Etag))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	s.Require().Equal("Wed, 21 Oct 2015 07:28:00 GMT", res.Header.Get(httpheaders.LastModified))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	s.Require().Empty(res.Header.Get(httpheaders.LastModified))
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	data := s.TestData.Read("test1.png")
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, lastModified)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Equal(lastModified, modifiedSince)
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, lastModified)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(304, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedDisabled() {
	data := s.TestData.Read("test1.png")
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, recentTimestamp)

	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		s.NoError(err)
		s.True(fileLastModified.Before(parsedModifiedSince))
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, recentTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(304, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	data := s.TestData.Read("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, oldTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true
	data := s.TestData.Read("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		s.NoError(err)
		s.True(fileLastModified.After(parsedModifiedSince))
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, oldTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvg() {
	config.AlwaysRasterizeSvg = true

	res := s.GET("/unsafe/rs:fill:40:40/plain/local:///test1.svg")

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithEnforceAvif() {
	config.AlwaysRasterizeSvg = true
	s.Config().Options.EnforceWebp = true

	res := s.GET("/unsafe/plain/local:///test1.svg", http.Header{"Accept": []string{"image/webp"}})

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/webp", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgDisabled() {
	config.AlwaysRasterizeSvg = false
	s.Config().Options.EnforceWebp = true

	res := s.GET("/unsafe/plain/local:///test1.svg")

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithFormat() {
	config.AlwaysRasterizeSvg = true
	config.SkipProcessingFormats = []imagetype.Type{imagetype.SVG}

	res := s.GET("/unsafe/plain/local:///test1.svg@svg")

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestMaxSrcFileSizeGlobal() {
	s.Config().Security.DefaultOptions.MaxSrcFileSize = 1

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)

	s.Require().Equal(422, res.StatusCode)
}

func TestProcessingHandler(t *testing.T) {
	suite.Run(t, new(ProcessingHandlerTestSuite))
}
