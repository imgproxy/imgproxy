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
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
)

const (
	testContainer = "test"
	testObject    = "foo/test.png"
)

type SwiftTestSuite struct {
	suite.Suite
	server       *swifttest.SwiftServer
	storage      storage.Reader
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

	c := generichttp.NewDefaultConfig()
	c.IgnoreSslVerification = true

	trans, err := generichttp.New(false, &c)
	s.Require().NoError(err)

	s.storage, err = New(s.T().Context(), &config, trans)
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
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/not-here.png", "")
	s.Require().NoError(err)
	s.Require().Equal(404, response.Status)
}

func (s *SwiftTestSuite) TestRoundTripReturns404WhenContainerNotFound() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "invalid", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(404, response.Status)
}

func (s *SwiftTestSuite) TestRoundTripWithETagEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *SwiftTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
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

func (s *SwiftTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *SwiftTestSuite) TestRoundTripWithLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *SwiftTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
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

func (s *SwiftTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func TestSwiftTransport(t *testing.T) {
	suite.Run(t, new(SwiftTestSuite))
}
