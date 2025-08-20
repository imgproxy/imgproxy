package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/transport"
	"github.com/stretchr/testify/suite"
)

const (
	testDataPath = "../../testdata"
)

type StreamerTestSuite struct {
	suite.Suite
	ts      *httptest.Server
	factory *Factory
}

func (s *StreamerTestSuite) SetupSuite() {
	config.Reset()
	config.AllowLoopbackSourceAddresses = true

	s.ts = httptest.NewServer(http.FileServer(http.Dir(testDataPath)))
}

func (s *StreamerTestSuite) TearDownSuite() {
	config.Reset()
	s.ts.Close()
}

func (s *StreamerTestSuite) SetupTest() {
	tr, err := transport.NewTransport()
	s.Require().NoError(err)

	fetcher, err := imagefetcher.NewFetcher(tr, imagefetcher.NewConfigFromEnv())
	s.Require().NoError(err)

	s.factory = New(NewConfigFromEnv(), headerwriter.NewConfigFromEnv(), fetcher)
}

func (s *StreamerTestSuite) TestStreamer() {
	const testFilePath = "/test1.jpg"

	// Read expected output from test data
	expected, err := os.ReadFile(filepath.Join(testDataPath, testFilePath))
	s.Require().NoError(err)

	// Prepare HTTP request and response recorder
	req := httptest.NewRequest("GET", testFilePath, nil)
	req.Header.Set(httpheaders.AcceptEncoding, "gzip")

	// Override the test server handler to assert Accept-Encoding header
	s.ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the Accept-Encoding header is passed through from original request
		s.Equal("gzip", r.Header.Get(httpheaders.AcceptEncoding))
		http.ServeFile(w, r, filepath.Join(testDataPath, r.URL.Path))
	})

	po := &options.ProcessingOptions{
		Filename: "xxx", // Override Content-Disposition
	}

	rr := httptest.NewRecorder()

	p := StreamingParams{
		UserRequest:       req,
		ImageURL:          s.ts.URL + testFilePath,
		ReqID:             "test-req-id",
		ProcessingOptions: po,
	}

	err = s.factory.NewHandler(context.Background(), &p, rr).Execute(context.Background())
	s.Require().NoError(err)

	// Check response body
	respBody := rr.Body.Bytes()
	s.Require().Equal(expected, respBody)

	// Check that Content-Disposition header is set correctly
	s.Require().Equal("inline; filename=\"xxx.jpg\"", rr.Header().Get("Content-Disposition"))
}

func TestStreamer(t *testing.T) {
	suite.Run(t, new(StreamerTestSuite))
}
