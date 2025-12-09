package swift

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/testsuite"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	swiftStorage testutil.LazyObj[*swiftStorageWrapper]
}

func (s *ReaderTestSuite) SetupSuite() {
	s.ReaderSuite.SetupSuite()

	s.TestContainer = "test-container"
	s.TestObjectKey = "test-object.txt"

	// Initialize Swift storage
	s.swiftStorage, _ = NewLazySuiteStorage(s.Lazy())

	// Swift test storage returns 200 for range requests
	// We have to skip partial content checks
	s.SkipPartialContentChecks = true

	s.Storage, _ = testutil.NewLazySuiteObj(s,
		func() (storage.Reader, error) {
			return s.swiftStorage().Storage, nil
		},
	)
}

func (s *ReaderTestSuite) SetupTest() {
	// Recreate Swift blob for each test
	conn := s.swiftStorage().Connection()
	f, err := conn.ObjectCreate(
		s.T().Context(), s.TestContainer, s.TestObjectKey, true, "", "application/octet-stream", nil,
	)
	s.Require().NoError(err)
	n, err := f.Write(s.TestData)
	s.Require().Len(s.TestData, n)
	s.Require().NoError(err)
	f.Close()
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
