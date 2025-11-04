package s3

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/testsuite"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	s3Storage testutil.LazyObj[*s3StorageWrapper]
}

func (s *ReaderTestSuite) SetupSuite() {
	s.ReaderSuite.SetupSuite()

	s.TestContainer = "test-container"
	s.TestObjectKey = "test-object.txt"

	// Initialize S3 storage
	s.s3Storage, _ = NewLazySuiteStorage(s.Lazy())

	s.Storage, _ = testutil.NewLazySuiteObj(s,
		func() (storage.Reader, error) {
			return s.s3Storage().Storage, nil
		},
	)
}

func (s *ReaderTestSuite) SetupTest() {
	// Recreate S3 blob for each test using backend directly
	backend := s.s3Storage().Server().Backend()
	metadata := map[string]string{
		"Content-Type":  "application/octet-stream",
		"Last-Modified": time.Now().Format(http.TimeFormat),
	}
	_, err := backend.PutObject(s.TestContainer, s.TestObjectKey, metadata,
		bytes.NewReader(s.TestData), int64(len(s.TestData)), nil)
	s.Require().NoError(err)
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
