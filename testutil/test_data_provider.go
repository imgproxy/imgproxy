package testutil

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// TestDataFolderName is the name of the testdata directory
	TestDataFolderName = "testdata"
)

// TestDataProviderT is a function that returns a [testing.T]
type TestDataProviderT func() *testing.T

// TestDataProvider provides access to test data images
type TestDataProvider struct {
	path string
	t    TestDataProviderT
}

// NewTestDataProvider creates a new TestDataProvider
func NewTestDataProvider(t TestDataProviderT) *TestDataProvider {
	t().Helper()

	path, err := findProjectRoot()
	if err != nil {
		require.NoError(t(), err)
	}

	return &TestDataProvider{
		path: filepath.Join(path, TestDataFolderName),
		t:    t,
	}
}

// findProjectRoot finds the absolute path to the project root by looking for go.mod
func findProjectRoot() (string, error) {
	// Start from current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, this is our project root
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding go.mod
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// Root returns the absolute path to the testdata directory
func (p *TestDataProvider) Root() string {
	return p.path
}

// Path returns the absolute path to a file in the testdata directory
func (p *TestDataProvider) Path(parts ...string) string {
	allParts := append([]string{p.path}, parts...)
	return filepath.Join(allParts...)
}

// Read reads a test data file and returns it as bytes
func (p *TestDataProvider) Read(name string) []byte {
	p.t().Helper()

	data, err := os.ReadFile(p.Path(name))
	require.NoError(p.t(), err)
	return data
}

// Reader reads a test data file and returns it as imagedata.ImageData
func (p *TestDataProvider) Reader(name string) *bytes.Reader {
	return bytes.NewReader(p.Read(name))
}

// FileEqualsToReader compares the contents of a test data file with the contents of the given reader
func (p *TestDataProvider) FileEqualsToReader(name string, reader io.Reader) bool {
	expected := p.Reader(name)
	return ReadersEqual(p.t(), expected, reader)
}
