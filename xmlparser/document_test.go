package xmlparser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentParsing(t *testing.T) {
	testImagesPath, err := filepath.Abs("../testdata/test-images/svg-test-suite")
	require.NoError(t, err)

	err = filepath.Walk(testImagesPath, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-SVG files
		if filepath.Ext(path) != ".svg" {
			return nil
		}

		// Open the SVG file
		file, err := os.Open(path)
		require.NoError(t, err)
		defer file.Close()

		// Parse the document
		doc1, err := NewDocument(file)
		require.NoError(t, err, "Failed to parse SVG: %s", path)

		// Write the document back to a buffer
		buf := new(bytes.Buffer)
		_, err = doc1.WriteTo(buf)
		require.NoError(t, err, "Failed to write SVG: %s", path)

		// Parse the document again from the written buffer
		doc2, err := NewDocument(buf)
		require.NoError(t, err, "Failed to re-parse SVG: %s", path)

		// Ensure that the two documents are equivalent
		require.Equal(t, doc1, doc2, "Documents do not match after re-parsing: %s", path)

		return nil
	})

	require.NoError(t, err)
}
