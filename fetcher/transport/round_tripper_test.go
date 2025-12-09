package transport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
)

// mockStorage is a simple mock implementation of storage.Reader
type mockStorage struct {
	getObject func(
		ctx context.Context, reqHeader http.Header, bucket, key, query string,
	) (*storage.ObjectReader, error)
}

func (m *mockStorage) GetObject(
	ctx context.Context, reqHeader http.Header, bucket, key, query string,
) (*storage.ObjectReader, error) {
	if m.getObject == nil {
		return nil, nil
	}

	return m.getObject(ctx, reqHeader, bucket, key, query)
}

type RoundTripperTestSuite struct {
	suite.Suite
}

func (s *RoundTripperTestSuite) TestRoundTripperSuccess() {
	// Create mock storage that returns a successful response
	mock := &mockStorage{
		getObject: func(
			ctx context.Context, reqHeader http.Header, bucket, key, query string,
		) (*storage.ObjectReader, error) {
			s.Equal("test-bucket", bucket)
			s.Equal("test-key", key)
			s.Equal("version=123", query)

			headers := make(http.Header)
			headers.Set(httpheaders.ContentType, "image/png")
			headers.Set(httpheaders.Etag, "test-etag")

			body := io.NopCloser(strings.NewReader("test data"))
			return storage.NewObjectOK(headers, body), nil
		},
	}

	rt := NewRoundTripper(mock, "?")

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, EscapeURL("s3://test-bucket/test-key?version=123"), nil)
	s.Require().NoError(err)

	// Execute RoundTrip
	resp, err := rt.RoundTrip(req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	// Verify response
	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("image/png", resp.Header.Get(httpheaders.ContentType))
	s.Equal("test-etag", resp.Header.Get(httpheaders.Etag))

	// Read and verify body
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	s.Equal("test data", string(data))
}

func (s *RoundTripperTestSuite) TestRoundTripperNotFound() {
	// Create mock storage that returns 404
	mock := &mockStorage{
		getObject: func(
			ctx context.Context, reqHeader http.Header, bucket, key, query string,
		) (*storage.ObjectReader, error) {
			return storage.NewObjectNotFound("object not found"), nil
		},
	}

	rt := NewRoundTripper(mock, "?")

	req, err := http.NewRequest(http.MethodGet, "s3://bucket/key", nil)
	s.Require().NoError(err)

	resp, err := rt.RoundTrip(req)
	s.Require().NoError(err)
	s.Require().NotNil(resp)

	if resp.Body != nil {
		resp.Body.Close()
	}

	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func TestRoundTripper(t *testing.T) {
	suite.Run(t, new(RoundTripperTestSuite))
}
