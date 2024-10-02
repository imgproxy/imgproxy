package azure

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type AzureTestSuite struct {
	suite.Suite

	server       *httptest.Server
	transport    http.RoundTripper
	etag         string
	lastModified time.Time
}

func (s *AzureTestSuite) SetupSuite() {
	data := make([]byte, 32)

	logrus.SetOutput(os.Stdout)

	config.IgnoreSslVerification = true

	s.etag = "testetag"
	s.lastModified, _ = time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")

	s.server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		s.Equal("/test/foo/test.png", r.URL.Path)

		rw.Header().Set("Etag", s.etag)
		rw.Header().Set("Last-Modified", s.lastModified.Format(http.TimeFormat))
		rw.WriteHeader(200)
		rw.Write(data)
	}))

	config.ABSEnabled = true
	config.ABSEndpoint = s.server.URL
	config.ABSName = "testname"
	config.ABSKey = "dGVzdGtleQ=="

	var err error
	s.transport, err = New()
	s.Require().NoError(err)
}

func (s *AzureTestSuite) TearDownSuite() {
	s.server.Close()
	config.IgnoreSslVerification = false
}

func (s *AzureTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get("ETag"))
}

func (s *AzureTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithLastModifiedDisabledReturns200() {
	config.LastModifiedEnabled = false
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}

func (s *AzureTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}
func TestAzureTransport(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
