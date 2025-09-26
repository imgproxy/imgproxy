package fs

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

type FsTestSuite struct {
	suite.Suite

	transport http.RoundTripper
	etag      string
	modTime   time.Time
}

func (s *FsTestSuite) SetupSuite() {
	wd, err := os.Getwd()
	s.Require().NoError(err)

	fsRoot := filepath.Join(wd, "..", "..", "..", "testdata")

	fi, err := os.Stat(filepath.Join(fsRoot, "test1.png"))
	s.Require().NoError(err)

	s.etag = BuildEtag("/test1.png", fi)
	s.modTime = fi.ModTime()
	s.transport, _ = New(&Config{Root: fsRoot})
}

func (s *FsTestSuite) TestRoundTripWithETagEnabled() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get(httpheaders.Etag))
}
func (s *FsTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithLastModifiedEnabledReturns200() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.modTime.Format(http.TimeFormat), response.Header.Get(httpheaders.LastModified))
}

func (s *FsTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.modTime.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *FsTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	request, _ := http.NewRequest("GET", "local:///test1.png", nil)
	request.Header.Set(httpheaders.IfModifiedSince, s.modTime.Add(-time.Minute).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func TestFSTransport(t *testing.T) {
	suite.Run(t, new(FsTestSuite))
}
