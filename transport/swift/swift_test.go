package swift

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	testContainer = "test"
	testObject    = "test.png"
)

type SwiftTestSuite struct {
	suite.Suite
	server *swifttest.SwiftServer
}

func (s *SwiftTestSuite) SetupSuite() {
	s.server, _ = swifttest.NewSwiftServer("localhost")

	config.Reset()

	config.SwiftAuthURL = s.server.AuthURL
	config.SwiftUsername = swifttest.TEST_ACCOUNT
	config.SwiftAPIKey = swifttest.TEST_ACCOUNT
	config.SwiftAuthVersion = 1

	s.setupTestFile()
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

	b, err := f.Write(data)
	assert.Greater(t, b, 100)
	assert.Nil(t, err)
}

func (s *SwiftTestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *SwiftTestSuite) TestNew() {
	transport, err := New()

	assert.Nil(s.T(), err, "failed to set up transport")
	assert.IsType(s.T(), transport, transport)
}

func (s *SwiftTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "swift:///test/test.png", nil)

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "swift:///test/test.png", nil)

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 200, response.StatusCode)
	assert.Equal(s.T(), "e27ca34142be8e55220e44155c626cd0", response.Header.Get("ETag"))
}

func (s *SwiftTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift:///test/test.png", nil)
	request.Header.Set("If-None-Match", "e27ca34142be8e55220e44155c626cd0")

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *SwiftTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "swift:///test/test.png", nil)
	request.Header.Set("If-None-Match", "foobar")

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, response.StatusCode)
}

func TestSwiftTransport(t *testing.T) {
	suite.Run(t, new(SwiftTestSuite))
}
