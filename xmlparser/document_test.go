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
		err = doc1.WriteTo(buf)
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

func TestEntityReplacement(t *testing.T) {
	svgData := []byte(`<?xml version="1.0"?>
<!DOCTYPE svg [
	<!ENTITY myEntity1 "EntityValue1">
	<!ENTITY myEntity2 'EntityValue2'>
]>
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
	<text x="10" y="20" arg1="Value with &myEntity1;" arg2='&myEntity2;'>
		&myEntity1; and &myEntity2;
	</text>
	<style>
		<![CDATA[
			.textClass { content: "&myEntity1;"; }
		]]>
	</style>
</svg>`)

	expectedData := []byte(`<?xml version="1.0"?>
<!DOCTYPE svg [
	<!ENTITY myEntity1 "EntityValue1">
	<!ENTITY myEntity2 'EntityValue2'>
]>
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
	<text x="10" y="20" arg1="Value with EntityValue1" arg2="EntityValue2">
		EntityValue1 and EntityValue2
	</text>
	<style>
		<![CDATA[
			.textClass { content: "&myEntity1;"; }
		]]>
	</style>
</svg>`)

	doc, err := NewDocument(bytes.NewReader(svgData))
	require.NoError(t, err)

	doc.ReplaceEntities()

	var buf bytes.Buffer
	err = doc.WriteTo(&buf)
	require.NoError(t, err)

	require.Equal(t, string(expectedData), buf.String())
}

func BenchmarkDocumentParsing(b *testing.B) {
	testImagesPath, err := filepath.Abs("../testdata/test-images/svg-test-suite")
	if err != nil {
		b.Fatal(err)
	}

	samples := [][]byte{}

	err = filepath.Walk(testImagesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			b.Fatal(err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-SVG files
		if filepath.Ext(path) != ".svg" {
			return nil
		}

		// Read SVG file
		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}

		samples = append(samples, data)
		return nil
	})

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for b.Loop() {
		for _, sample := range samples {
			_, err := NewDocument(bytes.NewReader(sample))
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
