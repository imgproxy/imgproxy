package gcs

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type GCSTestSuite struct {
	suite.Suite

	server       *fakestorage.Server
	transport    http.RoundTripper
	etag         string
	lastModified time.Time
}

func (s *GCSTestSuite) SetupSuite() {
	noAuth = true

	// s.etag = "testetag"
	s.lastModified, _ = time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")

	port, err := getFreePort()
	require.Nil(s.T(), err)

	s.server, err = fakestorage.NewServerWithOptions(fakestorage.Options{
		Scheme:     "http",
		Port:       uint16(port),
		PublicHost: fmt.Sprintf("localhost:%d", port),
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "test",
					Name:       "foo/test.png",
					// Etag:       s.etag,
					Updated: s.lastModified,
				},
				Content: make([]byte, 32),
			},
		},
	})
	require.Nil(s.T(), err)

	obj, err := s.server.GetObject("test", "foo/test.png")
	require.Nil(s.T(), err)
	s.etag = obj.Etag

	config.GCSEnabled = true
	config.GCSEndpoint = s.server.PublicURL() + "/storage/v1/"

	s.transport, err = New()
	require.Nil(s.T(), err)
}

func (s *GCSTestSuite) TearDownSuite() {
	s.server.Stop()
}

func (s *GCSTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.etag, response.Header.Get("ETag"))
}

func (s *GCSTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithLastModifiedDisabledReturns200() {
	config.LastModifiedEnabled = false
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}
func (s *GCSTestSuite) TestRoundTripWithLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.lastModified.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}
func (s *GCSTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}
func (s *GCSTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Sicne", s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}
func TestGCSTransport(t *testing.T) {
	suite.Run(t, new(GCSTestSuite))
}
