package swift

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
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

	config.Reset()

	config.SwiftAuthURL = s.server.AuthURL
	config.SwiftUsername = swifttest.TEST_ACCOUNT
	config.SwiftAPIKey = swifttest.TEST_ACCOUNT
	config.SwiftAuthVersion = 1

	s.setupTestFile()

	var err error
	s.transport, err = New()
	require.Nil(s.T(), err, "failed to initialize swift transport")
}

func (s *SwiftTestSuite) setupTestFile() {
	t := s.T()
	c := &swift.Connection{
		UserName:    config.SwiftUsername,
		ApiKey:      config.SwiftAPIKey,
		AuthUrl:     config.SwiftAuthURL,
		AuthVersion: config.SwiftAuthVersion,
	}

	ctx := context.Background()

	err := c.Authenticate(ctx)
	require.Nil(t, err, "failed to authenticate with test server")

	err = c.ContainerCreate(ctx, testContainer, nil)
	require.Nil(t, err, "failed to create container")

	f, err := c.ObjectCreate(ctx, testContainer, testObject, true, "", "image/png", nil)
	require.Nil(t, err, "failed to create object")

	defer f.Close()

	data := make([]byte, 32)

	n, err := f.Write(data)
	require.Equal(t, n, len(data))
	require.Nil(t, err)

	f.Close()
	// The Etag is written on file close; but Last-Modified is only available when we get the object again.
	_, h, err := c.Object(ctx, testContainer, testObject)
	require.Nil(t, err)
	s.etag = h["Etag"]
	s.lastModified, err = time.Parse(http.TimeFormat, h["Date"])
	require.Nil(t, err)
}

func (s *SwiftTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *SwiftTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripReturns404WhenObjectNotFound() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "swift://test/foo/not-here.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 404, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripReturns404WhenContainerNotFound() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "swift://invalid/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 404, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.etag, response.Header.Get("ETag"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithLastModifiedDisabledReturns200() {
	config.LastModifiedEnabled = false
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.lastModified.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func TestSwiftTransport(t *testing.T) {
	suite.Run(t, new(SwiftTestSuite))
}
