package gcs

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type GCSTestSuite struct {
	suite.Suite

	server       *fakestorage.Server
	storage      storage.Reader
	etag         string
	lastModified time.Time
	data         []byte
}

func (s *GCSTestSuite) SetupSuite() {
	// s.etag = "testetag"
	s.lastModified, _ = time.Parse(http.TimeFormat, "Wed, 21 Oct 2015 07:28:00 GMT")

	port, err := getFreePort()
	s.Require().NoError(err)

	s.data = make([]byte, 32)
	_, err = rand.Read(s.data)
	s.Require().NoError(err)

	s.server, err = fakestorage.NewServerWithOptions(fakestorage.Options{
		Scheme:     "http",
		Port:       uint16(port),
		PublicHost: fmt.Sprintf("localhost:%d", port),
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "test",
					Name:       "foo/test.png",
					// Etag:       s.etag,
					Updated: s.lastModified,
				},
				Content: s.data,
			},
		},
	})
	s.Require().NoError(err)

	obj, err := s.server.GetObject("test", "foo/test.png")
	s.Require().NoError(err)
	s.etag = obj.Etag

	config := NewDefaultConfig()
	config.Endpoint = s.server.PublicURL() + "/storage/v1/"

	c := generichttp.NewDefaultConfig()
	c.IgnoreSslVerification = true

	trans, err := generichttp.New(false, &c)
	s.Require().NoError(err)

	s.storage, err = New(&config, trans, false)
	s.Require().NoError(err)
}

func (s *GCSTestSuite) TearDownSuite() {
	s.server.Stop()
}

func (s *GCSTestSuite) TestRoundTripWithETagEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *GCSTestSuite) TestRoundTripWithIfNoneMatchReturns304() {
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

func (s *GCSTestSuite) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *GCSTestSuite) TestRoundTripWithLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}
func (s *GCSTestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
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

func (s *GCSTestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *GCSTestSuite) TestRoundTripWithRangeReturns206() {
	ctx := s.T().Context()

	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.Range, "bytes=10-19")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")

	s.Require().NoError(err)

	s.Require().Equal(http.StatusPartialContent, response.Status)
	s.Require().Equal(fmt.Sprintf("bytes 10-19/%d", 32), response.Headers.Get(httpheaders.ContentRange))
	s.Require().Equal("10", response.Headers.Get(httpheaders.ContentLength))
	s.Require().NotNil(response.Body)

	defer response.Body.Close()

	d, err := io.ReadAll(response.Body)
	s.Require().NoError(err)

	s.Require().Equal(d, s.data[10:20])
}

func TestGCSTransport(t *testing.T) {
	suite.Run(t, new(GCSTestSuite))
}
