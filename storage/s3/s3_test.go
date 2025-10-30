package s3

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
)

type S3TestSuite struct {
	suite.Suite

	server       *httptest.Server
	storage      storage.Reader
	etag         string
	lastModified time.Time
	data         []byte
}

func (s *S3TestSuite) SetupSuite() {
	backend := s3mem.New()
	faker := gofakes3.New(backend, gofakes3.WithIntegrityCheck(false))
	s.server = httptest.NewServer(faker.Server())

	config := NewDefaultConfig()
	config.Endpoint = s.server.URL

	os.Setenv("AWS_REGION", "eu-central-1")
	os.Setenv("AWS_ACCESS_KEY_ID", "Foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "Bar")

	c := generichttp.NewDefaultConfig()
	c.IgnoreSslVerification = true

	trans, err := generichttp.New(false, &c)
	s.Require().NoError(err)

	s.storage, err = New(&config, trans)
	s.Require().NoError(err)

	err = backend.CreateBucket("test")
	s.Require().NoError(err)

	svc := s.storage.(*Storage).defaultClient
	s.Require().NotNil(svc)
	s.Require().IsType(&s3.Client{}, svc)

	client := svc.(*s3.Client)

	s.data = make([]byte, 32)
	_, err = rand.Read(s.data)
	s.Require().NoError(err)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Body:   bytes.NewReader(s.data),
		Bucket: aws.String("test"),
		Key:    aws.String("foo/test.png"),
	})
	s.Require().NoError(err)

	obj, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String("test"),
		Key:    aws.String("foo/test.png"),
	})
	s.Require().NoError(err)
	defer obj.Body.Close()

	s.etag = *obj.ETag
	s.lastModified = *obj.LastModified
}

func (s *S3TestSuite) TearDownSuite() {
	s.server.Close()
}

func (s *S3TestSuite) TestRoundTripWithETagEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.etag, response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *S3TestSuite) TestRoundTripWithIfNoneMatchReturns304() {
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

func (s *S3TestSuite) TestRoundTripWithUpdatedETagReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, s.etag+"_wrong")

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *S3TestSuite) TestRoundTripWithLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *S3TestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
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

func (s *S3TestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.storage.GetObject(ctx, reqHeader, "test", "foo/test.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

func (s *S3TestSuite) TestRoundTripWithRangeReturns206() {
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

	// NOTE: err would contain CRC error, which is the limitation of s3 fake server
	d, _ := io.ReadAll(response.Body)

	s.Require().Equal(d, s.data[10:20])
}

func TestS3Transport(t *testing.T) {
	suite.Run(t, new(S3TestSuite))
}
