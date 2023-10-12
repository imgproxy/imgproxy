package fs

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/config"
)

type FsTestSuite struct {
	suite.Suite

	transport http.RoundTripper
	etag      string
	modTime   time.Time
}

func (s *FsTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	require.Nil(s.T(), err)

	config.LocalFileSystemRoot = filepath.Join(wd, "..", "..", "testdata")

	fi, err := os.Stat(filepath.Join(config.LocalFileSystemRoot, "test1.png"))
	require.Nil(s.T(), err)

	s.etag = BuildEtag("/test1.png", fi)
	s.modTime = fi.ModTime()
	s.transport = New()
}

func (s *FsTestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.etag, response.Header.Get("ETag"))
}
func (s *FsTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}
func (s *FsTestSuite) TestRoundTripWithLastModifiedDisabledReturns200() {
	config.LastModifiedEnabled = false
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithLastModifiedEnabledReturns200() {
	config.LastModifiedEnabled = true
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 200, response.StatusCode)
	require.Equal(s.T(), s.modTime.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}

func (s *FsTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set("If-Modified-Since", s.modTime.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusNotModified, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set("If-Modified-Since", s.modTime.Add(-time.Minute).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	require.Nil(s.T(), err)
	require.Equal(s.T(), http.StatusOK, response.StatusCode)
}
func TestS3Transport(t *testing.T) {
	suite.Run(t, new(FsTestSuite))
}
