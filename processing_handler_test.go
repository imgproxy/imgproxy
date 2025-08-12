package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/svg"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type ProcessingHandlerTestSuite struct {
	suite.Suite

	router *router.Router
}

func (s *ProcessingHandlerTestSuite) SetupSuite() {
	config.Reset()

	wd, err := os.Getwd()
	s.Require().NoError(err)

	s.T().Setenv("IMGPROXY_LOCAL_FILESYSTEM_ROOT", filepath.Join(wd, "/testdata"))
	s.T().Setenv("IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT", "0")

	err = initialize()
	s.Require().NoError(err)

	logrus.SetOutput(io.Discard)

	s.router = buildRouter()
}

func (s *ProcessingHandlerTestSuite) TeardownSuite() {
	shutdown()
	logrus.SetOutput(os.Stdout)
}

func (s *ProcessingHandlerTestSuite) SetupTest() {
	// We don't need config.LocalFileSystemRoot anymore as it is used
	// only during initialization
	config.Reset()
	config.AllowLoopbackSourceAddresses = true
}

func (s *ProcessingHandlerTestSuite) send(path string, header ...http.Header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rw := httptest.NewRecorder()

	if len(header) > 0 {
		req.Header = header[0]
	}

	s.router.ServeHTTP(rw, req)

	return rw
}

func (s *ProcessingHandlerTestSuite) readTestFile(name string) []byte {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(wd, "testdata", name))
	s.Require().NoError(err)

	return data
}

func (s *ProcessingHandlerTestSuite) readTestImageData(name string) imagedata.ImageData {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	data, err := os.ReadFile(filepath.Join(wd, "testdata", name))
	s.Require().NoError(err)

	imgdata, err := imagedata.NewFromBytes(data)
	s.Require().NoError(err)

	return imgdata
}

func (s *ProcessingHandlerTestSuite) readImageData(imgdata imagedata.ImageData) []byte {
	data, err := io.ReadAll(imgdata.Reader())
	s.Require().NoError(err)
	return data
}

func (s *ProcessingHandlerTestSuite) sampleETagData(imgETag string) (string, imagedata.ImageData, http.Header, string) {
	poStr := "rs:fill:4:4"

	po := options.NewProcessingOptions()
	po.ResizingType = options.ResizeFill
	po.Width = 4
	po.Height = 4

	imgdata := s.readTestImageData("test1.png")
	headers := make(http.Header)

	if len(imgETag) != 0 {
		headers.Set(httpheaders.Etag, imgETag)
	}

	var h etag.Handler

	h.SetActualProcessingOptions(po)
	h.SetActualImageData(imgdata, headers)
	return poStr, imgdata, headers, h.GenerateActualETag()
}

func (s *ProcessingHandlerTestSuite) TestRequest() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get("Content-Type"))

	format, err := imagetype.Detect(res.Body)

	s.Require().NoError(err)
	s.Require().Equal(imagetype.PNG, format)
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationFailure() {
	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(403, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationSuccess() {
	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}

	rw := s.send("/My9d3xq_PYpVHsPrCyww0Kh1w5KZeZhIlWhsa4az1TI/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
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
			expectedError:  false,
		},
		{
			name:           "match http URL with wildcard in hostname single level",
			allowedSources: []string{"local://", "http://*.mycdn.dev/"},
			requestPath:    "/unsafe/plain/http://a-1.mycdn.dev/lorem/ipsum.jpg",
			expectedError:  false,
		},
		{
			name:           "match http URL with wildcard in hostname multiple levels",
			allowedSources: []string{"local://", "http://*.mycdn.dev/"},
			requestPath:    "/unsafe/plain/http://a-1.b-2.mycdn.dev/lorem/ipsum.jpg",
			expectedError:  false,
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
			exps := make([]*regexp.Regexp, len(tc.allowedSources))
			for i, pattern := range tc.allowedSources {
				exps[i] = configurators.RegexpFromPattern(pattern)
			}
			config.AllowedSources = exps

			rw := s.send(tc.requestPath)
			res := rw.Result()

			if tc.expectedError {
				s.Require().Equal(404, res.StatusCode)
			} else {
				s.Require().Equal(200, res.StatusCode)
			}
		})
	}
}

func (s *ProcessingHandlerTestSuite) TestSourceNetworkValidation() {
	data := s.readTestFile("test1.png")

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer server.Close()

	var rw *httptest.ResponseRecorder

	u := fmt.Sprintf("/unsafe/rs:fill:4:4/plain/%s/test1.png", server.URL)

	rw = s.send(u)
	s.Require().Equal(200, rw.Result().StatusCode)

	config.AllowLoopbackSourceAddresses = false
	rw = s.send(u)
	s.Require().Equal(404, rw.Result().StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSourceFormatNotSupported() {
	vips.DisableLoadSupport(imagetype.PNG)
	defer vips.ResetLoadSupport()

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestResultingFormatNotSupported() {
	vips.DisableSaveSupport(imagetype.PNG)
	defer vips.ResetSaveSupport()

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	res := rw.Result()

	s.Require().Equal(422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingConfig() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected := s.readTestImageData("test1.png")

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingPO() {
	rw := s.send("/unsafe/rs:fill:4:4/skp:png/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected := s.readTestImageData("test1.png")

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSameFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected := s.readTestImageData("test1.png")

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingDifferentFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@jpg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected := s.readTestImageData("test1.png")

	s.Require().False(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSVG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.svg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected, err := svg.Sanitize(s.readTestImageData("test1.svg"))
	s.Require().NoError(err)

	s.Require().True(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestNotSkipProcessingSVGToJPG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.svg@jpg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)

	expected := s.readTestImageData("test1.svg")

	s.Require().False(testutil.ReadersEqual(s.T(), expected.Reader(), res.Body))
}

func (s *ProcessingHandlerTestSuite) TestErrorSavingToSVG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@svg")
	res := rw.Result()

	s.Require().Equal(422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughCacheControl() {
	config.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Cache-Control", "max-age=1234, public")
		rw.Header().Set("Expires", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	s.Require().Equal("max-age=1234, public", res.Header.Get("Cache-Control"))
	s.Require().Empty(res.Header.Get("Expires"))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughExpires() {
	config.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Expires", time.Now().Add(1239*time.Second).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	// Use regex to allow some delay
	s.Require().Regexp("max-age=123[0-9], public", res.Header.Get("Cache-Control"))
	s.Require().Empty(res.Header.Get("Expires"))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughDisabled() {
	config.CacheControlPassthrough = false

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Cache-Control", "max-age=1234, public")
		rw.Header().Set("Expires", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	s.Require().NotEqual("max-age=1234, public", res.Header.Get("Cache-Control"))
	s.Require().Empty(res.Header.Get("Expires"))
}

func (s *ProcessingHandlerTestSuite) TestETagDisabled() {
	config.ETagEnabled = false

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Empty(res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqNoIfNotModified() {
	config.ETagEnabled = true

	poStr, _, headers, etag := s.sampleETagData("loremipsumdolor")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", headers.Get(httpheaders.Etag))
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL))
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataNoIfNotModified() {
	config.ETagEnabled = true

	poStr, imgdata, _, etag := s.sampleETagData("")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(s.readImageData(imgdata))
	}))
	defer ts.Close()

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL))
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqMatch() {
	config.ETagEnabled = true

	poStr, _, headers, etag := s.sampleETagData(`"loremipsumdolor"`)

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Equal(headers.Get(httpheaders.Etag), r.Header.Get(httpheaders.IfNoneMatch))

		rw.WriteHeader(304)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", etag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	s.Require().Equal(304, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataMatch() {
	config.ETagEnabled = true

	poStr, imgdata, _, etag := s.sampleETagData("")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(s.readImageData(imgdata))
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", etag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	s.Require().Equal(304, res.StatusCode)
	s.Require().Equal(etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, headers, actualETag := s.sampleETagData(`"loremipsumdolor"`)
	_, _, _, expectedETag := s.sampleETagData(`"loremipsum"`)

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Equal(`"loremipsum"`, r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", headers.Get(httpheaders.Etag))
		rw.WriteHeader(200)
		rw.Write(s.readImageData(imgdata))
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(actualETag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, _, actualETag := s.sampleETagData("")
	// Change the data hash
	expectedETag := actualETag[:strings.IndexByte(actualETag, '/')] + "/Dasdbefj"

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(s.readImageData(imgdata))
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(actualETag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagProcessingOptionsNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, headers, actualETag := s.sampleETagData("")
	// Change the processing options hash
	expectedETag := "abcdefj" + actualETag[strings.IndexByte(actualETag, '/'):]

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Empty(r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", headers.Get(httpheaders.Etag))
		rw.WriteHeader(200)
		rw.Write(s.readImageData(imgdata))
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal(actualETag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	s.Require().Equal("Wed, 21 Oct 2015 07:28:00 GMT", res.Header.Get("Last-Modified"))
}

func (s *ProcessingHandlerTestSuite) TestLastModifiedDisabled() {
	config.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	s.Require().Empty(res.Header.Get("Last-Modified"))
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedDisabled() {
	config.LastModifiedEnabled = false
	data := s.readTestFile("test1.png")
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-Modified-Since", lastModified)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		s.Equal(lastModified, modifiedSince)
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-Modified-Since", lastModified)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(304, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedDisabled() {
	data := s.readTestFile("test1.png")
	config.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", recentTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get("If-Modified-Since")
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		s.NoError(err)
		s.True(fileLastModified.Before(parsedModifiedSince))
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", recentTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(304, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedDisabled() {
	config.LastModifiedEnabled = false
	data := s.readTestFile("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		s.Empty(modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", oldTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	data := s.readTestFile("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get("If-Modified-Since")
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		s.NoError(err)
		s.True(fileLastModified.After(parsedModifiedSince))
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", oldTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvg() {
	config.AlwaysRasterizeSvg = true

	rw := s.send("/unsafe/rs:fill:40:40/plain/local:///test1.svg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/png", res.Header.Get("Content-Type"))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithEnforceAvif() {
	config.AlwaysRasterizeSvg = true
	config.EnforceWebp = true

	rw := s.send("/unsafe/plain/local:///test1.svg", http.Header{"Accept": []string{"image/webp"}})
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/webp", res.Header.Get("Content-Type"))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgDisabled() {
	config.AlwaysRasterizeSvg = false
	config.EnforceWebp = true

	rw := s.send("/unsafe/plain/local:///test1.svg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get("Content-Type"))
}

func (s *ProcessingHandlerTestSuite) TestAlwaysRasterizeSvgWithFormat() {
	config.AlwaysRasterizeSvg = true
	config.SkipProcessingFormats = []imagetype.Type{imagetype.SVG}
	rw := s.send("/unsafe/plain/local:///test1.svg@svg")
	res := rw.Result()

	s.Require().Equal(200, res.StatusCode)
	s.Require().Equal("image/svg+xml", res.Header.Get("Content-Type"))
}

func (s *ProcessingHandlerTestSuite) TestMaxSrcFileSizeGlobal() {
	config.MaxSrcFileSize = 1

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	s.Require().Equal(422, res.StatusCode)
}

func TestProcessingHandler(t *testing.T) {
	suite.Run(t, new(ProcessingHandlerTestSuite))
}
