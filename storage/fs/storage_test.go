package fs_test

import (
	"context"

	"github.com/imgproxy/imgproxy/v4/storage/fs"
	"github.com/imgproxy/imgproxy/v4/testutil"
)

// LazySuiteStorage is a lazy object that provides FS storage for tests
type LazySuiteStorage = testutil.LazyObj[*fs.Storage]

// NewLazySuiteStorage creates a lazy FS Storage object for use in test suites
// The tmpDir parameter specifies the root directory for the filesystem storage
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
	tmpDir string,
) (testutil.LazyObj[*fs.Storage], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*fs.Storage, error) {
			config := fs.NewDefaultConfig()
			config.Root = tmpDir

			storage, err := fs.New(&config)
			if err != nil {
				return nil, err
			}

			return storage, nil
		},
	)
}
