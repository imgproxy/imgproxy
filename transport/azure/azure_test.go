package azure

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type AzureTestSuite struct {
	suite.Suite

	server    *httptest.Server
	transport http.RoundTripper
	etag      string
}

func (s *AzureTestSuite) SetupSuite() {
	data := make([]byte, 32)

	logrus.SetOutput(os.Stdout)

	s.etag = "testetag"

	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		require.Equal(s.T(), "/test/foo/test.png", r.URL.Path)

		rw.Header().Set("Etag", s.etag)
		rw.WriteHeader(200)
		rw.Write(data)
	}))

	config.ABSEnabled = true
	config.ABSEndpoint = s.server.URL
	config.ABSName = "testname"
	config.ABSKey = "dGVzdGtleQ=="

	var err error
	s.transport, err = New()
	require.Nil(s.T(), err)
}

func (s *AzureTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *AzureTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.etag, response.Header.Get("ETag"))
}

func (s *AzureTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *AzureTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "abs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func TestAzureTransport(t *testing.T) {
	suite.Run(t, new(AzureTestSuite))
}
