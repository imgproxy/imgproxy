package gcs

import (
	"testing"
	"time"

	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/testsuite"
	"github.com/imgproxy/imgproxy/v3/testutil"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	gcsStorage testutil.LazyObj[*gcsStorageWrapper]
}

func (s *ReaderTestSuite) SetupSuite() {
	s.ReaderSuite.SetupSuite()

	s.TestContainer = "test-container"
	s.TestObjectKey = "test-object.txt"

	// Prepare GCS storage with initial objects
	gcsInitialObjects := []fakestorage.Object{
		{
			ObjectAttrs: fakestorage.ObjectAttrs{
				BucketName: s.TestContainer,
				Name:       s.TestObjectKey,
				Updated:    time.Now(),
			},
			Content: s.TestData,
		},
	}

	// Initialize GCS storage
	s.gcsStorage, _ = NewLazySuiteStorage(s.Lazy(), gcsInitialObjects)

	s.Storage, _ = testutil.NewLazySuiteObj(s,
		func() (storage.Reader, error) {
			return s.gcsStorage().Storage, nil
		},
	)
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
