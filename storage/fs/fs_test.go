package fs

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type FsTestSuite struct {
	suite.Suite

	storage storage.Reader
	etag    string
	modTime time.Time
}

func (s *FsTestSuite) SetupSuite() {
	tdp := testutil.NewTestDataProvider(s.T)
	fsRoot := tdp.Root()

	fi, err := os.Stat(filepath.Join(fsRoot, "test1.png"))
	s.Require().NoError(err)

	s.etag = buildEtag("/test1.png", fi)
	s.modTime = fi.ModTime()

	s.storage, _ = New(&Config{Root: fsRoot}, "?")
}

func (s *FsTestSuite) TestRoundTripWithETagEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}
func (s *FsTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag)

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

func (s *FsTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *FsTestSuite) TestRoundTripWithLastModifiedEnabledReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.modTime.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *FsTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.modTime.Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

func (s *FsTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.modTime.Add(-time.Minute).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "", "test1.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func TestFSTransport(t *testing.T) {
	suite.Run(t, new(FsTestSuite))
}
