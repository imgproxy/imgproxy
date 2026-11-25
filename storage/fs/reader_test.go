package fs_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/imgproxy/imgproxy/v4/httpheaders"
	"github.com/imgproxy/imgproxy/v4/storage"
	"github.com/imgproxy/imgproxy/v4/storage/fs"
	"github.com/imgproxy/imgproxy/v4/storage/testsuite"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/stretchr/testify/suite"
)

type ReaderTestSuite struct {
	testsuite.ReaderSuite

	fsStorage testutil.LazyObj[*fs.Storage]
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

type trackingFile struct {
	http.File

	statErr error
	closed  bool
}

func (f *trackingFile) Stat() (os.FileInfo, error) {
	if f.statErr != nil {
		return nil, f.statErr
	}

	return f.File.Stat()
}

func (f *trackingFile) Close() error {
	f.closed = true
	return f.File.Close()
}

type trackingFS struct {
	http.FileSystem

	statErr error
	last    *trackingFile
}

func (t *trackingFS) Open(name string) (http.File, error) {
	f, err := t.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	t.last = &trackingFile{File: f, statErr: t.statErr}
	return t.last, nil
}

func (s *ReaderTestSuite) TestFileCloseLifecycle() {
	s.Require().NoError(os.MkdirAll(filepath.Join(s.tmpDir(), "subdir"), 0750))

	s.Run("Directory", func() {
		tfs := &trackingFS{FileSystem: http.Dir(s.tmpDir())}
		st := fs.NewWithFileSystem(tfs)

		obj, err := st.GetObject(s.T().Context(), http.Header{}, "", "subdir", "")
		s.Require().NoError(err)
		s.Require().Equal(http.StatusNotFound, obj.Status)
		s.Require().True(tfs.last.closed)
	})

	s.Run("StatError", func() {
		tfs := &trackingFS{FileSystem: http.Dir(s.tmpDir()), statErr: errors.New("stat failed")}
		st := fs.NewWithFileSystem(tfs)

		_, err := st.GetObject(s.T().Context(), http.Header{}, "", s.TestObjectKey, "")
		s.Require().ErrorIs(err, tfs.statErr)
		s.Require().True(tfs.last.closed)
	})

	s.Run("OK", func() {
		tfs := &trackingFS{FileSystem: http.Dir(s.tmpDir())}
		st := fs.NewWithFileSystem(tfs)

		obj, err := st.GetObject(s.T().Context(), http.Header{}, "", s.TestObjectKey, "")
		s.Require().NoError(err)
		s.Require().Equal(http.StatusOK, obj.Status)
		s.Require().False(tfs.last.closed)

		s.Require().NoError(obj.Body.Close())
		s.Require().True(tfs.last.closed)
	})

	s.Run("PartialContent", func() {
		tfs := &trackingFS{FileSystem: http.Dir(s.tmpDir())}
		st := fs.NewWithFileSystem(tfs)

		header := http.Header{httpheaders.Range: []string{"bytes=0-3"}}
		obj, err := st.GetObject(s.T().Context(), header, "", s.TestObjectKey, "")
		s.Require().NoError(err)
		s.Require().Equal(http.StatusPartialContent, obj.Status)
		s.Require().False(tfs.last.closed)

		s.Require().NoError(obj.Body.Close())
		s.Require().True(tfs.last.closed)
	})
}

func TestReader(t *testing.T) {
	suite.Run(t, new(ReaderTestSuite))
}
