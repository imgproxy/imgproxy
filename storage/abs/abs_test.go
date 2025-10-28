package azure

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/storage"
)

type AbsTest struct {
	suite.Suite

	server       *httptest.Server // TODO: use testutils.TestServer
	storage      storage.Reader
	etag         string
	lastModified time.Time
}

func (s *AbsTest) SetupSuite() {
	data := make([]byte, 32)

	logger.Mute()

	s.etag = "testetag"
	s.lastModified, _ = time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")

	s.server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Equal("/test/foo/test.png", r.URL.Path)

		rw.Header().Set(httpheaders.Etag, s.etag)
		rw.Header().Set(httpheaders.LastModified, s.lastModified.Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(data)
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

func (s *AbsTest) TearDownSuite() {
	s.server.Close()
	logger.Unmute()
}

func (s *AbsTest) TestRoundTripWithETag() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTest) TestRoundTripWithIfNoneMatchReturns304() {
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

func (s *AbsTest) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTest) TestRoundTripWithLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *AbsTest) TestRoundTripWithIfModifiedSinceReturns304() {
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

func (s *AbsTest) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func TestAzureTransport(t *testing.T) {
	suite.Run(t, new(AbsTest))
}
