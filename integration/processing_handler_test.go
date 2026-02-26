package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing/svg"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/stretchr/testify/suite"
)

// ProcessingHandlerTestSuite is a test suite for testing image processing handler
type ProcessingHandlerTestSuite struct {
	Suite
}

func (s *ProcessingHandlerTestSuite) SetupTest() {
	s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = true
}

func (s *ProcessingHandlerTestSuite) SetupSubTest() {
	// We use t.Run() a lot, so we need to reset lazy objects at the beginning of each subtest
	s.ResetLazyObjects()
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationFailure() {
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
			s.Config().Security.Keys = [][]byte{[]byte("test-key")}
			s.Config().Security.Salts = [][]byte{[]byte("test-salt")}

			res := s.GET(tc.url)
			defer res.Body.Close()

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
			s.Config().Security.AllowedSources = make([]*regexp.Regexp, len(tc.allowedSources))
			for i, pattern := range tc.allowedSources {
				s.Config().Security.AllowedSources[i] = env.RegexpFromPattern(pattern)
			}

			res := s.GET(tc.requestPath)
			defer res.Body.Close()

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
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer server.Close()

	url := fmt.Sprintf("/unsafe/rs:fill:4:4/plain/%s/test1.png", server.URL)

	// We wrap this in a subtest to reset s.router()
	s.Run("AllowLoopbackSourceAddressesTrue", func() {
		s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = true
		res := s.GET(url)
		defer res.Body.Close()

		s.Require().Equal(http.StatusOK, res.StatusCode)
	})

	s.Run("AllowLoopbackSourceAddressesFalse", func() {
		s.Config().Fetcher.Transport.HTTP.AllowLoopbackSourceAddresses = false
		res := s.GET(url)
		defer res.Body.Close()

		s.Require().Equal(http.StatusNotFound, res.StatusCode)
	})
}

func (s *ProcessingHandlerTestSuite) TestSourceFormatNotSupported() {
	vips.DisableLoadSupport(imagetype.PNG)
	defer vips.ResetLoadSupport()

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestResultingFormatNotSupported() {
	vips.DisableSaveSupport(imagetype.PNG)
	defer vips.ResetSaveSupport()

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingConfig() {
	s.Config().Processing.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingPO() {
	res := s.GET("/unsafe/rs:fill:4:4/skp:png/plain/local:///test1.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSameFormat() {
	s.Config().Processing.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingDifferentFormat() {
	s.Config().Processing.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@jpg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().False(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSVG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.svg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	data := imagedata.NewFromBytesWithFormat(imagetype.SVG, s.TestData.Read("test1.svg"))

	cfg := svg.NewDefaultConfig()
	svg := svg.New(&cfg)

	expected, err := svg.Process(&options.Options{}, data)
	s.Require().NoError(err)

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestNotSkipProcessingSVGToJPG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.svg@jpg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().False(s.TestData.FileEqualsToReader("test1.svg", res.Body))
}

func (s *ProcessingHandlerTestSuite) TestErrorSavingToSVG() {
	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png@svg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusUnprocessableEntity, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughCacheControl() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.CacheControl, "max-age=1234, public")
		rw.Header().Set(httpheaders.Expires, time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("max-age=1234, public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughExpires() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.Expires, time.Now().Add(1239*time.Second).UTC().Format(http.TimeFormat))
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	// Use regex to allow some delay
	s.Require().Regexp("max-age=123[0-9], public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughDisabled() {
	s.Config().Server.ResponseWriter.CacheControlPassthrough = false

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.CacheControl, "max-age=1234, public")
		rw.Header().Set(httpheaders.Expires, time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().NotEqual("max-age=1234, public", res.Header.Get(httpheaders.CacheControl))
	s.Require().Empty(res.Header.Get(httpheaders.Expires))
}

func (s *ProcessingHandlerTestSuite) TestETagDisabled() {
	s.Config().Handlers.Processing.ETagEnabled = false

	res := s.GET("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
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
	defer res.Body.Close()

	s.Require().Equal(http.StatusNotModified, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get(httpheaders.Etag))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Equal("Wed, 21 Oct 2015 07:28:00 GMT", res.Header.Get(httpheaders.LastModified))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Empty(res.Header.Get(httpheaders.LastModified))
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	data := s.TestData.Read("test1.png")
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, lastModified)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Equal(lastModified, modifiedSince)
		rw.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, lastModified)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusNotModified, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedBuster() {
	buster := time.Now()

	s.Config().Handlers.Processing.LastModifiedEnabled = true
	s.Config().Handlers.Processing.LastModifiedBuster = buster

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.LastModified, "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Equal(buster.Format(http.TimeFormat), res.Header.Get(httpheaders.LastModified))
}

func (s *ProcessingHandlerTestSuite) EtagBusterTest() {
	s.Config().Handlers.Processing.ETagEnabled = true
	s.Config().Handlers.Processing.ETagBuster = "buster"

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(httpheaders.Etag, `"loremipsumdolor"`)
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Equal(`"buster/bG9yZW1pcHN1bWRvbG9y"`, res.Header.Get(httpheaders.Etag))
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedDisabled() {
	data := s.TestData.Read("test1.png")
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, recentTimestamp)

	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedEnabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = true
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		s.NoError(err)
		s.True(fileLastModified.Before(parsedModifiedSince))
		rw.WriteHeader(http.StatusNotModified)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, recentTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusNotModified, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedDisabled() {
	s.Config().Handlers.Processing.LastModifiedEnabled = false
	data := s.TestData.Read("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get(httpheaders.IfModifiedSince)
		s.Empty(modifiedSince)
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, oldTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
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
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set(httpheaders.IfModifiedSince, oldTimestamp)
	res := s.GET(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvg() {
	s.Config().Processing.AlwaysRasterizeSvg = true

	res := s.GET("/unsafe/rs:fill:40:40/plain/local:///test1.svg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithEnforceWebP() {
	s.Config().Processing.AlwaysRasterizeSvg = true
	s.Config().ClientFeatures.EnforceWebp = true

	res := s.GET("/unsafe/plain/local:///test1.svg", http.Header{"Accept": []string{"image/webp"}})
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("image/webp", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgDisabled() {
	s.Config().Processing.AlwaysRasterizeSvg = false
	s.Config().ClientFeatures.EnforceWebp = true

	res := s.GET("/unsafe/plain/local:///test1.svg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithFormat() {
	s.Config().Processing.AlwaysRasterizeSvg = true
	s.Config().Processing.SkipProcessingFormats = []imagetype.Type{imagetype.SVG}

	res := s.GET("/unsafe/plain/local:///test1.svg@svg")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get(httpheaders.ContentType))
}

func (s *ProcessingHandlerTestSuite) TestMaxSrcFileSizeGlobal() {
	s.Config().Security.MaxSrcFileSize = 1

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.Write(s.TestData.Read("test1.png"))
	}))
	defer ts.Close()

	res := s.GET("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	defer res.Body.Close()

	s.Require().Equal(422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestRawOption() {
	res := s.GET("/unsafe/raw:1/plain/local:///test1.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().True(s.TestData.FileEqualsToReader("test1.png", res.Body))
}

func (s *ProcessingHandlerTestSuite) computeDist(url string, sourceHash *testutil.ImageHash) int {
	res := s.GET(url)
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	s.Require().NoError(err)

	processedImg, err := testutil.LoadImage(bytes.NewReader(body))
	s.Require().NoError(err)

	processedHash, err := testutil.NewImageHash(processedImg, testutil.HashTypePerception)
	s.Require().NoError(err)

	dist, err := sourceHash.Distance(processedHash)
	s.Require().NoError(err)

	return dist
}

func (s *ProcessingHandlerTestSuite) TestMaxBytes() {
	sourceData := s.TestData.Read("test-images/jpg/jpg.jpg")
	sourceSize := len(sourceData)

	mb := sourceSize / 2
	s.Require().Greater(sourceSize, mb, "Source image must be larger than mb for the test")

	res := s.GET(fmt.Sprintf("/unsafe/mb:%d/plain/local:///test-images/jpg/jpg.jpg@jpg", mb))
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	body, err := io.ReadAll(res.Body)

	s.Require().NoError(err)
	s.Require().LessOrEqual(len(body), mb)
}

func (s *ProcessingHandlerTestSuite) TestQualitySettings() {
	// Load source image and compute its hash
	sourceImg, err := testutil.LoadImage(bytes.NewReader(s.TestData.Read("test-images/jpg/jpg.jpg")))
	s.Require().NoError(err)

	sourceHash, err := testutil.NewImageHash(sourceImg, testutil.HashTypePerception)
	s.Require().NoError(err)

	// Set config quality to 99
	s.Config().Processing.Quality = 99

	// Test that high quality (99) or bypassed quality results in identical image
	s.Run("q_99", func() {
		dist := s.computeDist("/unsafe/q:99/plain/local:///test-images/jpg/jpg.jpg@jpg", sourceHash)
		s.Require().Equal(0, dist)
	})

	s.Run("no_q", func() {
		dist := s.computeDist("/unsafe/plain/local:///test-images/jpg/jpg.jpg@jpg", sourceHash)
		s.Require().Equal(0, dist)
	})

	s.Run("q_1_fq_1", func() {
		dist := s.computeDist("/unsafe/q:1/plain/local:///test-images/jpg/jpg.jpg@jpg", sourceHash)
		s.Require().NotEqual(0, dist)

		dist = s.computeDist("/unsafe/fq:jpg:1/plain/local:///test-images/jpg/jpg.jpg@jpg", sourceHash)
		s.Require().NotEqual(0, dist)
	})
}

func (s *ProcessingHandlerTestSuite) TestPresetUsage() {
	s.Config().OptionsParser.Presets = []string{
		"default=pixelate:10",
	}

	res := s.GET("/unsafe/preset:default/plain/local:///geometry.png")
	defer res.Body.Close()

	s.Require().Equal(http.StatusOK, res.StatusCode)
	s.Require().False(s.TestData.FileEqualsToReader("geometry.png", res.Body))
}

func TestProcessingHandler(t *testing.T) {
	suite.Run(t, new(ProcessingHandlerTestSuite))
}
