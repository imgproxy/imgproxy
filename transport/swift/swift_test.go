package swift

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

const (
	testContainer = "test"
	testObject    = "foo/test.png"
)

type SwiftTestSuite struct {
	suite.Suite
	server    *swifttest.SwiftServer
	transport http.RoundTripper
	etag      string
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
	assert.Nil(s.T(), err, "failed to initialize swift transport")
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
	assert.Nil(t, err, "failed to authenticate with test server")

	err = c.ContainerCreate(ctx, testContainer, nil)
	assert.Nil(t, err, "failed to create container")

	f, err := c.ObjectCreate(ctx, testContainer, testObject, true, "", "image/png", nil)
	assert.Nil(t, err, "failed to create object")

	defer f.Close()

	wd, err := os.Getwd()
	assert.Nil(t, err)

	data, err := ioutil.ReadFile(filepath.Join(wd, "..", "..", "testdata", "test1.png"))
	assert.Nil(t, err, "failed to read testdata/test1.png")

	n, err := f.Write(data)
	assert.Equal(t, n, len(data))
	assert.Nil(t, err)

	f.Close()

	h, err := f.Headers()
	assert.Nil(t, err)
	s.etag = h["Etag"]
}

func (s *SwiftTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *SwiftTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, response.StatusCode)
	assert.Equal(s.T(), s.etag, response.Header.Get("ETag"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func TestSwiftTransport(t *testing.T) {
	suite.Run(t, new(SwiftTestSuite))
}
