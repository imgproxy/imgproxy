package fs

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/testutil"
)

// LazySuiteStorage is a lazy object that provides FS storage for tests
type LazySuiteStorage = testutil.LazyObj[*Storage]

// NewLazySuiteStorage creates a lazy FS Storage object for use in test suites
// The tmpDir parameter specifies the root directory for the filesystem storage
func NewLazySuiteStorage(
	l testutil.LazySuiteFrom,
	tmpDir string,
) (testutil.LazyObj[*Storage], context.CancelFunc) {
	return testutil.NewLazySuiteObj(
		l,
		func() (*Storage, error) {
			config := NewDefaultConfig()
			config.Root = tmpDir

			storage, err := New(&config)
			if err != nil {
				return nil, err
			}

			return storage, nil
		},
	)
}
