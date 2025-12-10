package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/testsuite"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	fsStorage testutil.LazyObj[*Storage]
	tmpDir    testutil.LazyObj[string]
}

func (s *ReaderTestSuite) SetupSuite() {
	s.ReaderSuite.SetupSuite()
	s.TestObjectKey = "test-object.txt"

	s.tmpDir, _ = testutil.NewLazySuiteObj(s,
		func() (string, error) {
			return s.T().TempDir(), nil
		})

	s.fsStorage, _ = NewLazySuiteStorage(s.Lazy(), s.tmpDir())
	s.Storage, _ = testutil.NewLazySuiteObj(s,
		func() (storage.Reader, error) {
			return s.fsStorage(), nil
		},
	)
}

func (s *ReaderTestSuite) SetupTest() {
	// Prepare FS storage - write test file directly
	testFile := filepath.Join(s.tmpDir(), s.TestObjectKey)

	err := os.MkdirAll(filepath.Dir(testFile), 0750)
	s.Require().NoError(err)

	err = os.WriteFile(testFile, s.TestData, 0600)
	s.Require().NoError(err)
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
