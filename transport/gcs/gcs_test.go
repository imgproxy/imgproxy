package gcs

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/transport/generichttp"
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
	s.Require().NoError(err)

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
	s.Require().NoError(err)

	obj, err := s.server.GetObject("test", "foo/test.png")
	s.Require().NoError(err)
	s.etag = obj.Etag

	config := NewDefaultConfig()
	config.Endpoint = s.server.PublicURL() + "/storage/v1/"

	tc := generichttp.NewDefaultConfig()
	tc.IgnoreSslVerification = true

	trans, gerr := generichttp.New(false, tc)
	s.Require().NoError(gerr)

	s.transport, err = New(config, trans)
	s.Require().NoError(err)
}

func (s *GCSTestSuite) TearDownSuite() {
	s.server.Stop()
}

func (s *GCSTestSuite) TestRoundTripWithETagEnabled() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get(httpheaders.Etag))
}

func (s *GCSTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithLastModifiedEnabled() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Header.Get(httpheaders.LastModified))
}
func (s *GCSTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *GCSTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	request, _ := http.NewRequest("GET", "gcs://test/foo/test.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}
func TestGCSTransport(t *testing.T) {
	suite.Run(t, new(GCSTestSuite))
}
