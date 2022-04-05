package swift

import (
	"context"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/ncw/swift/v2"
	"github.com/ncw/swift/v2/swifttest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

const (
	testContainer = "test"
	testObject    = "test.png"
)

func initTestServer() (*swifttest.SwiftServer, error) {
	server, err := swifttest.NewSwiftServer("localhost")

	config.SwiftAuthURL = server.AuthURL
	config.SwiftUsername = swifttest.TEST_ACCOUNT
	config.SwiftAPIKey = swifttest.TEST_ACCOUNT
	config.SwiftAuthVersion = 1

	return server, err
}

func setupTestFile(t *testing.T) {
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

	wd, err := os.Getwd()
	assert.Nil(t, err)

	data, err := ioutil.ReadFile(filepath.Join(wd, "..", "..", "testdata", "test1.png"))
	assert.Nil(t, err, "failed to read testdata/test1.png")

	b, err := f.Write(data)
	assert.Greater(t, b, 100)
	assert.Nil(t, err)
	f.Close()
}

func TestNew(t *testing.T) {
	server, err := initTestServer()
	assert.Nil(t, err, "failed to set up test server")
	defer server.Close()

	transport, err := New()

	assert.Nil(t, err, "failed to set up transport")
	assert.IsType(t, transport, transport)
}

func TestTransport_RoundTripWithETagDisabledReturns200(t *testing.T) {
	server, err := initTestServer()
	assert.Nil(t, err, "failed to set up test server")
	defer server.Close()

	setupTestFile(t)

	request, _ := http.NewRequest("GET", "swift:///swifttest/test/test.png", nil)

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
}

func TestTransport_RoundTripWithETagEnabled(t *testing.T) {
	config.ETagEnabled = true
	server, err := initTestServer()
	assert.Nil(t, err, "failed to set up test server")
	defer server.Close()

	setupTestFile(t)

	request, _ := http.NewRequest("GET", "swift:///swifttest/test/test.png", nil)

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "e27ca34142be8e55220e44155c626cd0", response.Header.Get("ETag"))
}

func TestTransport_RoundTripWithIfNoneMatchReturns304(t *testing.T) {
	config.ETagEnabled = true
	server, err := initTestServer()
	assert.Nil(t, err, "failed to set up test server")
	defer server.Close()

	setupTestFile(t)

	request, _ := http.NewRequest("GET", "swift:///swifttest/test/test.png", nil)
	request.Header.Set("If-None-Match", "e27ca34142be8e55220e44155c626cd0")

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNotModified, response.StatusCode)
}

func TestTransport_RoundTripWithUpdatedETagReturns200(t *testing.T) {
	config.ETagEnabled = true
	server, err := initTestServer()
	assert.Nil(t, err, "failed to set up test server")
	defer server.Close()

	setupTestFile(t)

	request, _ := http.NewRequest("GET", "swift:///swifttest/test/test.png", nil)
	request.Header.Set("If-None-Match", "foobar")

	transport, _ := New()

	response, err := transport.RoundTrip(request)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
}
