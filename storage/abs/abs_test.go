package abs

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/storage"
)

type AbsTestSuite struct {
	suite.Suite

	server       *httptest.Server // TODO: use testutils.TestServer
	storage      storage.Reader
	etag         string
	lastModified time.Time
	data         []byte
}

func (s *AbsTestSuite) SetupSuite() {
	s.data = make([]byte, 32)
	_, err := rand.Read(s.data)
	s.Require().NoError(err)

	logger.Mute()

	s.etag = "testetag"
	s.lastModified, _ = time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")

	s.server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Equal("/test/foo/test.png", r.URL.Path)

		// Azure client transforms "Range" header to "X-Ms-Range"
		rangeHeader := r.Header.Get("X-Ms-Range")
		if rangeHeader != "" {
			// Parse range header: "bytes=start-end"
			rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.Split(rangeStr, "-")
			if len(parts) == 2 {
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])
				if end >= len(s.data) {
					end = len(s.data) - 1
				}

				rw.Header().Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, len(s.data)))
				rw.Header().Set(httpheaders.ContentLength, fmt.Sprintf("%d", end-start+1))
				rw.WriteHeader(http.StatusPartialContent)
				rw.Write(s.data[start : end+1])
				return
			}
		}

		rw.Header().Set(httpheaders.Etag, s.etag)
		rw.Header().Set(httpheaders.LastModified, s.lastModified.Format(http.TimeFormat))

		rw.WriteHeader(200)
		rw.Write(s.data)
	}))

	config := NewDefaultConfig()
	config.Endpoint = s.server.URL
	config.Name = "testname"
	config.Key = "dGVzdGtleQ=="

	c := generichttp.NewDefaultConfig()
	c.IgnoreSslVerification = true

	trans, err := generichttp.New(false, &c)
	s.Require().NoError(err)

	s.storage, err = New(&config, trans)
	s.Require().NoError(err)
}

func (s *AbsTestSuite) TearDownSuite() {
	s.server.Close()
	logger.Unmute()
}

func (s *AbsTestSuite) TestRoundTripWithETag() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

func (s *AbsTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTestSuite) TestRoundTripWithLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

func (s *AbsTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTestSuite) TestRoundTripWithRangeReturns206() {
	ctx := s.T().Context()

	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.Range, "bytes=10-19")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")

	s.Require().NoError(err)

	s.Require().Equal(http.StatusPartialContent, response.Status)
	s.Require().Equal(fmt.Sprintf("bytes 10-19/%d", 32), response.Headers.Get(httpheaders.ContentRange))
	s.Require().Equal("10", response.Headers.Get(httpheaders.ContentLength))
	s.Require().NotNil(response.Body)

	defer response.Body.Close()

	d, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

	s.Require().Equal(d, s.data[10:20])
}

func TestAzureTransport(t *testing.T) {
	suite.Run(t, new(AbsTestSuite))
}
