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
)

type AzureTestSuite struct {
	suite.Suite

	server       *httptest.Server // TODO: use testutils.TestServer
	transport    http.RoundTripper
	etag         string
	lastModified time.Time
}

func (s *AzureTestSuite) SetupSuite() {
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

	tc := generichttp.NewDefaultConfig()
	tc.IgnoreSslVerification = true

	trans, gerr := generichttp.New(false, &tc, "?")
	s.Require().NoError(gerr)

	var err error
	s.transport, err = New(&config, trans, "?")
	s.Require().NoError(err)
}

func (s *AzureTestSuite) TearDownSuite() {
	s.server.Close()
	logger.Unmute()
}

func (s *AzureTestSuite) TestRoundTripWithETag() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get(httpheaders.Etag))
}

func (s *AzureTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithLastModifiedEnabled() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Header.Get(httpheaders.LastModified))
}

func (s *AzureTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func TestAzureTransport(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
