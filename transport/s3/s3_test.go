package s3

import (
	"bytes"
	"context"
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

	"github.com/imgproxy/imgproxy/v3/config"
)

type S3TestSuite struct {
	suite.Suite

	server       *httptest.Server
	transport    http.RoundTripper
	etag         string
	lastModified time.Time
}

func (s *S3TestSuite) SetupSuite() {
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	s.server = httptest.NewServer(faker.Server())

	config.S3Enabled = true
	config.S3Endpoint = s.server.URL

	os.Setenv("AWS_REGION", "eu-central-1")
	os.Setenv("AWS_ACCESS_KEY_ID", "Foo")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "Bar")

	var err error
	s.transport, err = New()
	s.Require().NoError(err)

	err = backend.CreateBucket("test")
	s.Require().NoError(err)

	svc := s.transport.(*transport).defaultClient
	s.Require().NotNil(svc)
	s.Require().IsType(&s3.Client{}, svc)

	client := svc.(*s3.Client)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Body:   bytes.NewReader(make([]byte, 32)),
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
	config.Reset()
}

func (s *S3TestSuite) TestRoundTripWithETagDisabledReturns200() {
	config.ETagEnabled = false
	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
}

func (s *S3TestSuite) TestRoundTripWithETagEnabled() {
	config.ETagEnabled = true
	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.etag, response.Header.Get("ETag"))
}

func (s *S3TestSuite) TestRoundTripWithIfNoneMatchReturns304() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *S3TestSuite) TestRoundTripWithUpdatedETagReturns200() {
	config.ETagEnabled = true

	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)
	request.Header.Set("If-None-Match", s.etag+"_wrong")

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func (s *S3TestSuite) TestRoundTripWithLastModifiedDisabledReturns200() {
	config.LastModifiedEnabled = false
	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
}

func (s *S3TestSuite) TestRoundTripWithLastModifiedEnabled() {
	config.LastModifiedEnabled = true
	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(200, response.StatusCode)
	s.Require().Equal(s.lastModified.Format(http.TimeFormat), response.Header.Get("Last-Modified"))
}

func (s *S3TestSuite) TestRoundTripWithIfModifiedSinceReturns304() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.StatusCode)
}

func (s *S3TestSuite) TestRoundTripWithUpdatedLastModifiedReturns200() {
	config.LastModifiedEnabled = true

	request, _ := http.NewRequest("GET", "s3://test/foo/test.png", nil)
	request.Header.Set("If-Modified-Since", s.lastModified.Add(-24*time.Hour).Format(http.TimeFormat))

	response, err := s.transport.RoundTrip(request)
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.StatusCode)
}

func TestS3Transport(t *testing.T) {
	suite.Run(t, new(S3TestSuite))
}
