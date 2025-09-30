package swift

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
)

const (
	testContainer = "test"
	testObject    = "foo/test.png"
)

type SwiftTestSuite struct {
	suite.Suite
	server       *swifttest.SwiftServer
	transport    http.RoundTripper
	etag         string
	lastModified time.Time
}

func (s *SwiftTestSuite) SetupSuite() {
	s.server, _ = swifttest.NewSwiftServer("localhost")

	config := NewDefaultConfig()

	config.AuthURL = s.server.AuthURL
	config.Username = swifttest.TEST_ACCOUNT
	config.APIKey = swifttest.TEST_ACCOUNT
	config.AuthVersion = 1

	s.setupTestFile(&config)

	tc := generichttp.NewDefaultConfig()
	tc.IgnoreSslVerification = true

	trans, gerr := generichttp.New(false, &tc)
	s.Require().NoError(gerr)

	var err error
	s.transport, err = New(&config, trans, "?")
	s.Require().NoError(err, "failed to initialize swift transport")
}

func (s *SwiftTestSuite) setupTestFile(config *Config) {
	c := &swift.Connection{
		UserName:    config.Username,
		ApiKey:      config.APIKey,
		AuthUrl:     config.AuthURL,
		AuthVersion: config.AuthVersion,
	}

	ctx := context.Background()

	err := c.Authenticate(ctx)
	s.Require().NoError(err, "failed to authenticate with test server")

	err = c.ContainerCreate(ctx, testContainer, nil)
	s.Require().NoError(err, "failed to create container")

	f, err := c.ObjectCreate(ctx, testContainer, testObject, true, "", "image/png", nil)
	s.Require().NoError(err, "failed to create object")

	defer f.Close()

	data := make([]byte, 32)

	n, err := f.Write(data)
	s.Require().Len(data, n)
	s.Require().NoError(err)

	f.Close()
	// The Etag is written on file close; but Last-Modified is only available when we get the object again.
	_, h, err := c.Object(ctx, testContainer, testObject)
	s.Require().NoError(err)
	s.etag = h["Etag"]
	s.lastModified, err = time.Parse(http.TimeFormat, h["Date"])
	s.Require().NoError(err)
}

func (s *SwiftTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *SwiftTestSuite) TestRoundTripReturns404WhenObjectNotFound() {
	request, _ := http.NewRequest("GET", "swift://test/foo/not-here.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(404, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripReturns404WhenContainerNotFound() {
	request, _ := http.NewRequest("GET", "swift://invalid/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(404, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithETagEnabled() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get("ETag"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithLastModifiedEnabled() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func TestSwiftTransport(t *testing.T) {
	suite.Run(t, new(SwiftTestSuite))
}
