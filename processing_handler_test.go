package main

import (
	"bytes"
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

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/etag"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/svg"
	"github.com/imgproxy/imgproxy/v3/vips"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ProcessingHandlerTestSuite struct {
	suite.Suite

	router *router.Router
}

func (s *ProcessingHandlerTestSuite) SetupSuite() {
	config.Reset()

	wd, err := os.Getwd()
	require.Nil(s.T(), err)

	config.LocalFileSystemRoot = filepath.Join(wd, "/testdata")
	// Disable keep-alive to test connection restrictions
	config.ClientKeepAliveTimeout = 0

	err = initialize()
	require.Nil(s.T(), err)

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
	require.Nil(s.T(), err)

	data, err := os.ReadFile(filepath.Join(wd, "testdata", name))
	require.Nil(s.T(), err)

	return data
}

func (s *ProcessingHandlerTestSuite) readBody(res *http.Response) []byte {
	data, err := io.ReadAll(res.Body)
	require.Nil(s.T(), err)
	return data
}

func (s *ProcessingHandlerTestSuite) sampleETagData(imgETag string) (string, *imagedata.ImageData, string) {
	poStr := "rs:fill:4:4"

	po := options.NewProcessingOptions()
	po.ResizingType = options.ResizeFill
	po.Width = 4
	po.Height = 4

	imgdata := imagedata.ImageData{
		Type: imagetype.PNG,
		Data: s.readTestFile("test1.png"),
	}

	if len(imgETag) != 0 {
		imgdata.Headers = map[string]string{"ETag": imgETag}
	}

	var h etag.Handler

	h.SetActualProcessingOptions(po)
	h.SetActualImageData(&imgdata)
	return poStr, &imgdata, h.GenerateActualETag()
}

func (s *ProcessingHandlerTestSuite) TestRequest() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), "image/png", res.Header.Get("Content-Type"))

	meta, err := imagemeta.DecodeMeta(res.Body)

	require.Nil(s.T(), err)
	require.Equal(s.T(), imagetype.PNG, meta.Format())
	require.Equal(s.T(), 4, meta.Width())
	require.Equal(s.T(), 4, meta.Height())
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationFailure() {
	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 403, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSignatureValidationSuccess() {
	config.Keys = [][]byte{[]byte("test-key")}
	config.Salts = [][]byte{[]byte("test-salt")}

	rw := s.send("/My9d3xq_PYpVHsPrCyww0Kh1w5KZeZhIlWhsa4az1TI/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
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
		s.T().Run(tc.name, func(t *testing.T) {
			exps := make([]*regexp.Regexp, len(tc.allowedSources))
			for i, pattern := range tc.allowedSources {
				exps[i] = configurators.RegexpFromPattern(pattern)
			}
			config.AllowedSources = exps

			rw := s.send(tc.requestPath)
			res := rw.Result()

			if tc.expectedError {
				require.Equal(s.T(), 404, res.StatusCode)
			} else {
				require.Equal(s.T(), 200, res.StatusCode)
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
	fmt.Println(u)

	rw = s.send(u)
	require.Equal(s.T(), 200, rw.Result().StatusCode)

	config.AllowLoopbackSourceAddresses = false
	rw = s.send(u)
	require.Equal(s.T(), 404, rw.Result().StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSourceFormatNotSupported() {
	vips.DisableLoadSupport(imagetype.PNG)
	defer vips.ResetLoadSupport()

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestResultingFormatNotSupported() {
	vips.DisableSaveSupport(imagetype.PNG)
	defer vips.ResetSaveSupport()

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	res := rw.Result()

	require.Equal(s.T(), 422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingConfig() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected := s.readTestFile("test1.png")

	require.True(s.T(), bytes.Equal(expected, actual))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingPO() {
	rw := s.send("/unsafe/rs:fill:4:4/skp:png/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected := s.readTestFile("test1.png")

	require.True(s.T(), bytes.Equal(expected, actual))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSameFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected := s.readTestFile("test1.png")

	require.True(s.T(), bytes.Equal(expected, actual))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingDifferentFormat() {
	config.SkipProcessingFormats = []imagetype.Type{imagetype.PNG}

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@jpg")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected := s.readTestFile("test1.png")

	require.False(s.T(), bytes.Equal(expected, actual))
}

func (s *ProcessingHandlerTestSuite) TestSkipProcessingSVG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.svg")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected, err := svg.Satitize(&imagedata.ImageData{Data: s.readTestFile("test1.svg")})

	require.Nil(s.T(), err)

	require.True(s.T(), bytes.Equal(expected.Data, actual))
}

func (s *ProcessingHandlerTestSuite) TestNotSkipProcessingSVGToJPG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.svg@jpg")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)

	actual := s.readBody(res)
	expected := s.readTestFile("test1.svg")

	require.False(s.T(), bytes.Equal(expected, actual))
}

func (s *ProcessingHandlerTestSuite) TestErrorSavingToSVG() {
	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png@svg")
	res := rw.Result()

	require.Equal(s.T(), 422, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthrough() {
	config.CacheControlPassthrough = true

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Cache-Control", "fake-cache-control")
		rw.Header().Set("Expires", "fake-expires")
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	require.Equal(s.T(), "fake-cache-control", res.Header.Get("Cache-Control"))
	require.Equal(s.T(), "fake-expires", res.Header.Get("Expires"))
}

func (s *ProcessingHandlerTestSuite) TestCacheControlPassthroughDisabled() {
	config.CacheControlPassthrough = false

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Cache-Control", "fake-cache-control")
		rw.Header().Set("Expires", "fake-expires")
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send("/unsafe/rs:fill:4:4/plain/" + ts.URL)
	res := rw.Result()

	require.NotEqual(s.T(), "fake-cache-control", res.Header.Get("Cache-Control"))
	require.NotEqual(s.T(), "fake-expires", res.Header.Get("Expires"))
}

func (s *ProcessingHandlerTestSuite) TestETagDisabled() {
	config.ETagEnabled = false

	rw := s.send("/unsafe/rs:fill:4:4/plain/local:///test1.png")
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Empty(s.T(), res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqNoIfNotModified() {
	config.ETagEnabled = true

	poStr, imgdata, etag := s.sampleETagData("loremipsumdolor")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Empty(s.T(), r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", imgdata.Headers["ETag"])
		rw.WriteHeader(200)
		rw.Write(s.readTestFile("test1.png"))
	}))
	defer ts.Close()

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL))
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataNoIfNotModified() {
	config.ETagEnabled = true

	poStr, imgdata, etag := s.sampleETagData("")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Empty(s.T(), r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(imgdata.Data)
	}))
	defer ts.Close()

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL))
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqMatch() {
	config.ETagEnabled = true

	poStr, imgdata, etag := s.sampleETagData(`"loremipsumdolor"`)

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Equal(s.T(), imgdata.Headers["ETag"], r.Header.Get("If-None-Match"))

		rw.WriteHeader(304)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", etag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 304, res.StatusCode)
	require.Equal(s.T(), etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataMatch() {
	config.ETagEnabled = true

	poStr, imgdata, etag := s.sampleETagData("")

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Empty(s.T(), r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(imgdata.Data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", etag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 304, res.StatusCode)
	require.Equal(s.T(), etag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagReqNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, actualETag := s.sampleETagData(`"loremipsumdolor"`)
	_, _, expectedETag := s.sampleETagData(`"loremipsum"`)

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Equal(s.T(), `"loremipsum"`, r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", imgdata.Headers["ETag"])
		rw.WriteHeader(200)
		rw.Write(imgdata.Data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), actualETag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagDataNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, actualETag := s.sampleETagData("")
	// Change the data hash
	expectedETag := actualETag[:strings.IndexByte(actualETag, '/')] + "/Dasdbefj"

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Empty(s.T(), r.Header.Get("If-None-Match"))

		rw.WriteHeader(200)
		rw.Write(imgdata.Data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), actualETag, res.Header.Get("ETag"))
}

func (s *ProcessingHandlerTestSuite) TestETagProcessingOptionsNotMatch() {
	config.ETagEnabled = true

	poStr, imgdata, actualETag := s.sampleETagData("")
	// Change the processing options hash
	expectedETag := "abcdefj" + actualETag[strings.IndexByte(actualETag, '/'):]

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Empty(s.T(), r.Header.Get("If-None-Match"))

		rw.Header().Set("ETag", imgdata.Headers["ETag"])
		rw.WriteHeader(200)
		rw.Write(imgdata.Data)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-None-Match", expectedETag)

	rw := s.send(fmt.Sprintf("/unsafe/%s/plain/%s", poStr, ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
	require.Equal(s.T(), actualETag, res.Header.Get("ETag"))
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

	require.Equal(s.T(), "Wed, 21 Oct 2015 07:28:00 GMT", res.Header.Get("Last-Modified"))
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

	require.Equal(s.T(), "", res.Header.Get("Last-Modified"))
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedDisabled() {
	config.LastModifiedEnabled = false
	data := s.readTestFile("test1.png")
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		require.Equal(s.T(), "", modifiedSince)
		rw.WriteHeader(200)
		rw.Write(data)

	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-Modified-Since", lastModified)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
}
func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqExactMatchLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	lastModified := "Wed, 21 Oct 2015 07:28:00 GMT"
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		require.Equal(s.T(), lastModified, modifiedSince)
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	header := make(http.Header)
	header.Set("If-Modified-Since", lastModified)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 304, res.StatusCode)
}

func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedDisabled() {
	data := s.readTestFile("test1.png")
	config.LastModifiedEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		require.Equal(s.T(), modifiedSince, "")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", recentTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
}
func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareMoreRecentLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get("If-Modified-Since")
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		require.Nil(s.T(), err)
		require.True(s.T(), fileLastModified.Before(parsedModifiedSince))
		rw.WriteHeader(304)
	}))
	defer ts.Close()

	recentTimestamp := "Thu, 25 Feb 2021 01:45:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", recentTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 304, res.StatusCode)
}
func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedDisabled() {
	config.LastModifiedEnabled = false
	data := s.readTestFile("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		modifiedSince := r.Header.Get("If-Modified-Since")
		require.Equal(s.T(), modifiedSince, "")
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", oldTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
}
func (s *ProcessingHandlerTestSuite) TestModifiedSinceReqCompareTooOldLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	data := s.readTestFile("test1.png")
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileLastModified, _ := time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")
		modifiedSince := r.Header.Get("If-Modified-Since")
		parsedModifiedSince, err := time.Parse(http.TimeFormat, modifiedSince)
		require.Nil(s.T(), err)
		require.True(s.T(), fileLastModified.After(parsedModifiedSince))
		rw.WriteHeader(200)
		rw.Write(data)
	}))
	defer ts.Close()

	oldTimestamp := "Tue, 01 Oct 2013 17:31:00 GMT"

	header := make(http.Header)
	header.Set("If-Modified-Since", oldTimestamp)
	rw := s.send(fmt.Sprintf("/unsafe/plain/%s", ts.URL), header)
	res := rw.Result()

	require.Equal(s.T(), 200, res.StatusCode)
}
func TestProcessingHandler(t *testing.T) {
	suite.Run(t, new(ProcessingHandlerTestSuite))
}
