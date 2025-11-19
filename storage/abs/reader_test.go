package abs

import (
	"crypto/rand"
	"testing"

	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/testsuite"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

const (
	testDataSize = 128
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	absStorage testutil.LazyObj[*absStorageWrapper]
}

func (s *ReaderTestSuite) SetupSuite() {
	s.ReaderSuite.SetupSuite()

	// Generate random test data for content verification
	s.TestData = make([]byte, testDataSize)
	rand.Read(s.TestData)

	s.TestContainer = "test-container"
	s.TestObjectKey = "test-object.txt"

	// Initialize ABS storage
	s.absStorage, _ = NewLazySuiteStorage(s.Lazy())

	s.Storage, _ = testutil.NewLazySuiteObj(s,
		func() (storage.Reader, error) {
			return s.absStorage().Storage, nil
		},
	)
}

func (s *ReaderTestSuite) SetupTest() {
	// Recreate ABS blob for each test
	abs := s.absStorage().Client().ServiceClient().NewContainerClient(s.TestContainer).NewBlockBlobClient(s.TestObjectKey)
	_, err := abs.UploadBuffer(s.T().Context(), s.TestData, nil)
	s.Require().NoError(err)
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
