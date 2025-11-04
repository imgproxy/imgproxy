package testsuite

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

const (
	testDataSize = 128
)

type ReaderSuite struct {
	testutil.LazySuite

	Storage       testutil.LazyObj[storage.Reader]
	TestContainer string
	TestObjectKey string
	TestData      []byte

	SkipPartialContentChecks bool
}

func (s *ReaderSuite) SetupSuite() {
	// Generate random test data for content verification
	s.TestData = make([]byte, testDataSize)
	rand.Read(s.TestData)
}

// TestETagEnabled verifies that ETag header is returned in responses
func (s *ReaderSuite) TestETagEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().NotEmpty(response.Headers.Get(httpheaders.Etag))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

// TestIfNoneMatchReturns304 verifies that If-None-Match header causes 304 response when ETag matches
func (s *ReaderSuite) TestIfNoneMatchReturns304() {
	ctx := s.T().Context()

	// First, get the ETag
	reqHeader := make(http.Header)
	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	etag := response.Headers.Get(httpheaders.Etag)
	s.Require().NotEmpty(etag)
	response.Body.Close()

	// Now request with If-None-Match
	reqHeader = make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, etag)

	response, err = s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

// TestUpdatedETagReturns200 verifies that a wrong If-None-Match header returns 200
func (s *ReaderSuite) TestUpdatedETagReturns200() {
	ctx := s.T().Context()

	// First, get the ETag
	reqHeader := make(http.Header)
	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	etag := response.Headers.Get(httpheaders.Etag)
	s.Require().NotEmpty(etag)
	response.Body.Close()

	// Now request with wrong ETag
	reqHeader = make(http.Header)
	reqHeader.Set(httpheaders.IfNoneMatch, etag+"_wrong")

	response, err = s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

// TestLastModifiedEnabled verifies that Last-Modified header is returned in responses
func (s *ReaderSuite) TestLastModifiedEnabled() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	s.Require().NotEmpty(response.Headers.Get(httpheaders.LastModified))
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

// TestIfModifiedSinceReturns304 verifies that If-Modified-Since header causes 304 response when date matches
func (s *ReaderSuite) TestIfModifiedSinceReturns304() {
	ctx := s.T().Context()

	// First, get the Last-Modified time
	reqHeader := make(http.Header)
	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	lastModified := response.Headers.Get(httpheaders.LastModified)
	s.Require().NotEmpty(lastModified)
	response.Body.Close()

	// Now request with If-Modified-Since
	reqHeader = make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, lastModified)

	response, err = s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotModified, response.Status)

	if response.Body != nil {
		response.Body.Close()
	}
}

// TestUpdatedLastModifiedReturns200 verifies that an older If-Modified-Since header returns 200
func (s *ReaderSuite) TestUpdatedLastModifiedReturns200() {
	ctx := s.T().Context()

	// First, get the Last-Modified time
	reqHeader := make(http.Header)
	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(200, response.Status)
	lastModifiedStr := response.Headers.Get(httpheaders.LastModified)
	s.Require().NotEmpty(lastModifiedStr)
	response.Body.Close()

	lastModified, err := time.Parse(http.TimeFormat, lastModifiedStr)
	s.Require().NoError(err)

	// Now request with older If-Modified-Since
	reqHeader = make(http.Header)
	reqHeader.Set(httpheaders.IfModifiedSince, lastModified.Add(-time.Minute).Format(http.TimeFormat))

	response, err = s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusOK, response.Status)
	s.Require().NotNil(response.Body)

	response.Body.Close()
}

// TestRangeRequest verifies that Range header returns partial content
func (s *ReaderSuite) TestRangeRequest() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)
	reqHeader.Set(httpheaders.Range, "bytes=10-19")

	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, s.TestObjectKey, "")
	s.Require().NoError(err)

	if !s.SkipPartialContentChecks {
		s.Require().Equal(http.StatusPartialContent, response.Status)

		expectedRange := fmt.Sprintf("bytes 10-19/%d", len(s.TestData))
		s.Require().Equal(expectedRange, response.Headers.Get(httpheaders.ContentRange))
	}

	s.Require().Equal("10", response.Headers.Get(httpheaders.ContentLength))
	s.Require().NotNil(response.Body)

	// Read and verify the actual content (bytes 10-19 from testData)
	buf := make([]byte, 10)
	n, _ := response.Body.Read(buf)
	s.Require().Equal(10, n)
	s.Require().Equal(s.TestData[10:20], buf)

	response.Body.Close()
}

// TestObjectNotFound verifies that requesting a non-existent object returns 404
func (s *ReaderSuite) TestObjectNotFound() {
	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.Storage().GetObject(ctx, reqHeader, s.TestContainer, "nonexistent/object.png", "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotFound, response.Status)
}

// TestContainerNotFound verifies that requesting from a non-existent container returns 404
func (s *ReaderSuite) TestContainerNotFound() {
	if s.TestContainer == "" {
		s.T().Skip("Test container is blank: skipping test")
	}

	ctx := s.T().Context()
	reqHeader := make(http.Header)

	response, err := s.Storage().GetObject(ctx, reqHeader, "nonexistent-container", s.TestObjectKey, "")
	s.Require().NoError(err)
	s.Require().Equal(http.StatusNotFound, response.Status)
}
